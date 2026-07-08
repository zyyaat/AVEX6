package storage

import (
        "bytes"
        "encoding/json"
        "fmt"
        "io"
        "net/http"
        "os"
        "strings"
        "time"

        "github.com/google/uuid"
)

const sidecarEndpoint = "http://127.0.0.1:1106"

// PrivateObjectDir returns the private object directory path (e.g. "/bucket-id/.private").
func PrivateObjectDir() string {
        return os.Getenv("PRIVATE_OBJECT_DIR")
}

// Enabled reports whether object storage has been provisioned.
func Enabled() bool {
        return PrivateObjectDir() != ""
}

func parseObjectPath(path string) (bucket, object string, err error) {
        if !strings.HasPrefix(path, "/") {
                path = "/" + path
        }
        parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
        if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
                return "", "", fmt.Errorf("invalid object path: %s", path)
        }
        return parts[0], parts[1], nil
}

func signObjectURL(bucket, object, method string, ttlSec int) (string, error) {
        body := map[string]interface{}{
                "bucket_name": bucket,
                "object_name": object,
                "method":      method,
                "expires_at":  time.Now().Add(time.Duration(ttlSec) * time.Second).UTC().Format(time.RFC3339),
        }
        b, _ := json.Marshal(body)
        req, err := http.NewRequest("POST", sidecarEndpoint+"/object-storage/signed-object-url", bytes.NewReader(b))
        if err != nil {
                return "", err
        }
        req.Header.Set("Content-Type", "application/json")
        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
                return "", err
        }
        defer resp.Body.Close()
        if resp.StatusCode != 200 {
                data, _ := io.ReadAll(resp.Body)
                return "", fmt.Errorf("sign url failed (%d): %s", resp.StatusCode, string(data))
        }
        var out struct {
                SignedURL string `json:"signed_url"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
                return "", err
        }
        return out.SignedURL, nil
}

// GetUploadURL generates a presigned PUT URL under the private object dir, plus
// the normalized objectPath (e.g. "/objects/order-photos/<uuid>") to persist in the DB.
func GetUploadURL(category string) (uploadURL string, objectPath string, err error) {
        dir := PrivateObjectDir()
        if dir == "" {
                return "", "", fmt.Errorf("PRIVATE_OBJECT_DIR not set")
        }
        id := uuid.New().String()
        if category == "" {
                category = "uploads"
        }
        entityID := category + "/" + id
        fullPath := strings.TrimSuffix(dir, "/") + "/" + entityID
        bucket, object, err := parseObjectPath(fullPath)
        if err != nil {
                return "", "", err
        }
        url, err := signObjectURL(bucket, object, "PUT", 900)
        if err != nil {
                return "", "", err
        }
        return url, "/objects/" + entityID, nil
}

// GetDownloadURL generates a presigned GET URL for an objectPath previously returned by GetUploadURL.
func GetDownloadURL(objectPath string) (string, error) {
        if !strings.HasPrefix(objectPath, "/objects/") {
                return "", fmt.Errorf("invalid object path")
        }
        entityID := strings.TrimPrefix(objectPath, "/objects/")
        dir := PrivateObjectDir()
        if dir == "" {
                return "", fmt.Errorf("PRIVATE_OBJECT_DIR not set")
        }
        fullPath := strings.TrimSuffix(dir, "/") + "/" + entityID
        bucket, object, err := parseObjectPath(fullPath)
        if err != nil {
                return "", err
        }
        return signObjectURL(bucket, object, "GET", 3600)
}

// RegisterRoutes exposes GET /api/storage/objects/{path...} which redirects to a
// short-lived presigned download URL for the underlying private object.
func RegisterRoutes(mux *http.ServeMux) {
        mux.HandleFunc("GET /api/storage/objects/{path...}", func(w http.ResponseWriter, r *http.Request) {
                objectPath := "/objects/" + r.PathValue("path")
                url, err := GetDownloadURL(objectPath)
                if err != nil {
                        http.Error(w, "not found", http.StatusNotFound)
                        return
                }
                http.Redirect(w, r, url, http.StatusFound)
        })
}
