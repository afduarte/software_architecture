package main

func AuthRoutes(s *Server) {
	s.router.GET("/auth", noOp(s))
}
