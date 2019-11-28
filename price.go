package main

func PriceRoutes(s *Server) {
	s.router.GET("/price", noOp(s))
}
