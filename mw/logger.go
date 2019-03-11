package mw

import (
	"io"
	"os"
	"time"

	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

// Logger sends IIS Log File Format and updates Prometheus stats.

type (
	// LoggerConfig defines the config for Logger middleware.
	LoggerConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Output is a writer where logs in JSON format are written.
		// Optional. Default value os.Stdout.
		Output io.Writer
	}
)

var (
	// DefaultLoggerConfig is the default Logger middleware config.
	DefaultLoggerConfig = LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Output:  os.Stdout,
	}

	requestsHandled = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "apigw",
			Subsystem: "logger",
			Name:      "handled_total",
			Help:      "Counter of fully handled requests",
		}, []string{"clientid", "status"})

	requestsHandlingTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "apigw",
			Subsystem: "logger",
			Name:      "handling_duration_seconds",
			Help:      "Histogram of handling time of successful requests",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 13),
		}, []string{"method"})
)

func init() {
	prometheus.MustRegister(requestsHandled)
	prometheus.MustRegister(requestsHandlingTime)
}

// Logger returns a middleware that logs HTTP requests.
func Logger() echo.MiddlewareFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig returns a Logger middleware with config.
// See: `Logger()`.
func LoggerWithConfig(config LoggerConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}
			// invoke handler
			start := time.Now()
			err = next(c)
			if err != nil {
				c.Error(err)
			}
			stop := time.Now()

			// get values for later on

			delta := stop.Sub(start)
			req := c.Request()
			res := c.Response()

			clientID, ok := c.Get("ClientID").(string)
			if !ok {
				clientID = "-"
			}

			bytesIn := req.Header.Get(echo.HeaderContentLength)
			if bytesIn == "" {
				bytesIn = "0"
			}

			// Update Prometheus stats
			requestsHandled.WithLabelValues(clientID, strconv.FormatInt(int64(res.Status), 10)).Inc()
			requestsHandlingTime.WithLabelValues(req.Method).Observe(delta.Seconds())

			// Write log in IIS log format.
			//      IP address, user name, req.date and time,    format, service                mSec, req, resp, status,    method,          path,
			// 192.168.114.201,         -, 03/20/01,  7:55:20,   W3SVC2, SALES1, 172.21.13.45,  4502, 163, 3223,    200, 0,    GET, /DeptLogo.gif, -,
			// see https://msdn.microsoft.com/en-us/library/ms525807(v=vs.90).aspx
			fmt.Fprintf(config.Output, "%s,%s,%s,W3SVC,%s, -,%d,%s,%d,%d,0,%s,%s, -,\n",
				c.RealIP(),
				clientID,
				start.Format("01/02/06,15:04:05"),
				req.Host,
				//req.Host IP
				delta.Nanoseconds()/1000000,
				bytesIn,
				res.Size,
				res.Status,
				// hardcoded 0
				req.Method,
				req.URL.Path)

			return
		}
	}
}
