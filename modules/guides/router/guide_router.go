package router

import (
	"nimbus-service/modules/guides/handler"

	"github.com/go-chi/chi/v5"
)

// RegisterGuideRoutes sets up the routes for the guide module
func RegisterGuideRoutes(router chi.Router, h *handler.GuideHandler) {
	router.Route("/guides", func(r chi.Router) {
		r.Post("/", h.CreateGuide)                                  // POST /guides
		r.Get("/{guideID}", h.GetGuideByID)                         // GET /guides/{guideID}
		r.Get("/", h.ListGuides)                                    // GET /guides
		r.Put("/{guideID}", h.UpdateGuide)                          // PUT /guides/{guideID}
		r.Delete("/{guideID}", h.DeleteGuide)                       // DELETE /guides/{guideID}
		r.Get("/phone/{areaCode}/{phoneNumber}", h.GetGuideByPhone) // GET /guides/phone/{areaCode}/{phoneNumber}
	})
}
