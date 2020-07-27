package app

/*var NotFoundHandler = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		resp := response.New(false, "The resource was not found on our server")
		err := resp.Respond(w)
		if err != nil {
			// TODO: Log not panic
			panic(err)
		}

		next.ServeHTTP(w, r)
	})
}*/
