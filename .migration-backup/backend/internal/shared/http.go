package shared

import (
	"encoding/json"
	"net/http"
	"strings"
)

func WriteJSON(w http.ResponseWriter, s int, d interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	json.NewEncoder(w).Encode(d)
}

func WriteErr(w http.ResponseWriter, s int, m string) {
	WriteJSON(w, s, map[string]string{"error": m})
}

// ===== Middlewares =====

func AuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah == "" {
			WriteErr(w, 401, "غير مصرح")
			return
		}
		parts := strings.Split(ah, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteErr(w, 401, "صيغة خاطئة")
			return
		}
		c, err := VerifyJWT(parts[1])
		if err != nil {
			WriteErr(w, 401, "رمز غير صالح")
			return
		}
		ctx := ContextWithUser(r, c)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func OptionalAuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah != "" {
			parts := strings.Split(ah, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				if c, err := VerifyJWT(parts[1]); err == nil {
					r = r.WithContext(ContextWithUser(r, c))
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func AdminMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := GetUser(r)
		if c == nil || !c.Admin {
			WriteErr(w, 403, "غير مصرح - مطلوب مسؤول")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func DriverAuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah == "" {
			WriteErr(w, 401, "غير مصرح")
			return
		}
		parts := strings.Split(ah, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteErr(w, 401, "صيغة خاطئة")
			return
		}
		c, err := VerifyJWT(parts[1])
		if err != nil {
			WriteErr(w, 401, "رمز غير صالح")
			return
		}
		if !c.IsDriver {
			WriteErr(w, 403, "هذا المسار للمندوبين فقط")
			return
		}
		ctx := ContextWithUser(r, c)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func MerchantAuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah == "" {
			WriteErr(w, 401, "غير مصرح")
			return
		}
		parts := strings.Split(ah, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteErr(w, 401, "صيغة خاطئة")
			return
		}
		c, err := VerifyJWT(parts[1])
		if err != nil {
			WriteErr(w, 401, "رمز غير صالح")
			return
		}
		if !c.IsMerchant {
			WriteErr(w, 403, "هذا المسار للتجار فقط")
			return
		}
		ctx := ContextWithUser(r, c)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AgentAuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah == "" {
			WriteErr(w, 401, "غير مصرح")
			return
		}
		parts := strings.Split(ah, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteErr(w, 401, "صيغة خاطئة")
			return
		}
		c, err := VerifyJWT(parts[1])
		if err != nil {
			WriteErr(w, 401, "رمز غير صالح")
			return
		}
		if !c.IsAgent {
			WriteErr(w, 403, "هذا المسار لموظفي الدعم فقط")
			return
		}
		ctx := ContextWithUser(r, c)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
