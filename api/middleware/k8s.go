package middleware

func IsK8sPath(path string) bool {
	if path == "/api/v1/health" || path == "/api/v1/ready" {
		return true
	}
	return false
}
