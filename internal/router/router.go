package router

import (
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"

	"github.com/xmapst/osreapi/internal/router/api/v1/pool"
	"github.com/xmapst/osreapi/internal/router/api/v1/sys"
	"github.com/xmapst/osreapi/internal/router/api/v1/task"
	"github.com/xmapst/osreapi/internal/router/api/v1/task/step"
	"github.com/xmapst/osreapi/internal/router/api/v1/task/workspace"
	taskv2 "github.com/xmapst/osreapi/internal/router/api/v2/task"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/pkg/info"
	"github.com/xmapst/osreapi/pkg/logx"
)

func New() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		middleware.RealIP,
		middleware.NoCache,
		middleware.Heartbeat("/heartbeat"),
		//middleware.Compress(gzip.DefaultCompression),
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
			AllowedHeaders:   []string{"Origin", "Content-Length", "Content-Type"},
			AllowCredentials: false,
			MaxAge:           300,
		}),
		header,
		logger,
		recovery,
	)

	// debug pprof
	router.Mount("/debug", middleware.Profiler())

	// base
	router.Get("/version", version)
	router.Get("/healthyz", healthyz)
	router.Route("/api", func(r chi.Router) {
		// v1
		r.Route("/v1", func(r chi.Router) {
			// pool
			r.Route("/pool", func(r chi.Router) {
				r.Get("/", pool.Detail)
				r.Post("/", pool.Post)
			})
			// pty
			r.Route("/pty", func(r chi.Router) {
				r.Get("/", sys.PtyWs)
			})
			// task
			r.Route("/task", func(r chi.Router) {
				r.Get("/", task.List)
				r.Post("/", task.Post)
				r.Route("/{task}", func(r chi.Router) {
					r.Get("/", task.Get)
					r.Put("/", task.Manager)
					// 	workspace
					r.Route("/workspace", func(r chi.Router) {
						r.Get("/", workspace.Get)
						r.Post("/", workspace.Post)
						r.Delete("/", workspace.Delete)
					})
					// step
					r.Route("/step", func(r chi.Router) {
						r.Route("/{step}", func(r chi.Router) {
							r.Get("/", step.Log)
							r.Put("/", step.Manager)
						})
					})
				})
			})
		})
		// v2
		r.Route("/v2", func(r chi.Router) {
			// task
			r.Route("/task", func(r chi.Router) {
				r.Post("/", taskv2.Post)
			})
		})
	})

	// no method
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(errors.New("method not allowed")))
	})

	// no route
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(errors.New("the requested path does not exist")))
	})
	return router
}

func logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()
		defer func() {
			logx.Infoln(r.RemoteAddr, r.Method, r.Proto, ww.Status(), r.URL.String(), time.Since(t1), r.UserAgent())
		}()
		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}

func recovery(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				logx.Errorln(rvr, debug.Stack())
				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func header(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Server", "chi")
		w.Header().Add("X-Version", info.Version)
		w.Header().Add("X-Powered-By", info.UserEmail)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
