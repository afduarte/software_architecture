package main

func OrderRoutes(s *Server) {
	s.router.GET("/order", noOp(s))
}
