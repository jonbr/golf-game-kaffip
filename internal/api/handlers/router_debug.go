package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
)

const (
	colorReset   = "\033[0m"
	colorBlue    = "\033[34m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorCyan    = "\033[36m"
	colorMagenta = "\033[35m"
)

func PrintRoutes(r chi.Routes) {
	groups := make(map[string][]routeEntry)

	chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		group := extractGroup(route)

		mwNames := extractMiddlewareNames(middlewares)

		groups[group] = append(groups[group], routeEntry{
			Method:     method,
			Route:      route,
			Middleware: mwNames,
		})

		return nil
	})

	groupNames := make([]string, 0, len(groups))
	for g := range groups {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)

	fmt.Println()
	fmt.Println(colorCyan + "Registered Routes:" + colorReset)

	for _, group := range groupNames {
		fmt.Printf("\n%s%s%s\n", colorYellow, group, colorReset)
		fmt.Println(colorYellow + "--------------------------------------------------------" + colorReset)
		fmt.Printf("%s%-8s%s | %-30s | %s%s%s\n",
			colorYellow, "METHOD", colorReset,
			"ROUTE",
			colorYellow, "MIDDLEWARE", colorReset,
		)
		fmt.Println(colorYellow + "--------------------------------------------------------" + colorReset)

		routes := groups[group]
		sort.Slice(routes, func(i, j int) bool {
			return routes[i].Route < routes[j].Route
		})

		for _, entry := range routes {
			mw := "-"
			if len(entry.Middleware) > 0 {
				mw = strings.Join(entry.Middleware, ", ")
			}

			fmt.Printf("%s%-8s%s | %s%-30s%s | %s%s%s\n",
				colorGreen, entry.Method, colorReset,
				colorBlue, entry.Route, colorReset,
				colorMagenta, mw, colorReset,
			)
		}

		fmt.Println(colorYellow + "--------------------------------------------------------" + colorReset)
	}

	fmt.Println()
}

type routeEntry struct {
	Method     string
	Route      string
	Middleware []string
}

func extractGroup(route string) string {
	if route == "/" {
		return "/"
	}
	parts := strings.Split(route, "/")
	if len(parts) > 1 && parts[1] != "" {
		return "/" + parts[1]
	}
	return "/"
}

func extractMiddlewareNames(mws []func(http.Handler) http.Handler) []string {
	names := []string{}
	for _, mw := range mws {
		full := runtimeFuncName(mw)
		parts := strings.Split(full, ".")
		names = append(names, parts[len(parts)-1])
	}
	return names
}

func runtimeFuncName(i interface{}) string {
	full := fmt.Sprintf("%T", i)
	return full
}
