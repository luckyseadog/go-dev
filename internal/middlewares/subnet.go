package middlewares

import (
	"net"
	"net/http"
)

func SubnetMiddleware(trustedSubnet string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				next.ServeHTTP(w, r)

			} else {
				_, ipNet, err := net.ParseCIDR(trustedSubnet)
				if err != nil {
					http.Error(w, "SubnetMiddleware: Error with SubnetMiddleware var!", http.StatusBadRequest)
					return
				}

				IP := r.Header.Get("X-Real-IP")
				if !ipNet.Contains(net.ParseIP(IP)) {
					http.Error(w, "Invalid IP", http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
			}
		})
	}
}
