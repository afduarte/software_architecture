package main

func LoyaltyRoutes(s *Server) {
	s.router.GET("/loyalty", noOp(s))
}
