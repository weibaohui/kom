package kom

import (
    "net/http"
    "net/url"
    "time"

    "github.com/dgraph-io/ristretto/v2"
    "k8s.io/client-go/rest"
)

// RegisterParams carries registration-time options only. It is consumed during cluster registration
// and is NOT stored on Kubectl or ClusterInst for post-registration usage.
type RegisterParams struct {
    // rest.Config options
    ProxyURL      string
    ProxyFunc     func(*http.Request) (*url.URL, error)
    Timeout       time.Duration
    QPS           *float32
    Burst         *int
    UserAgent     string
    TLSInsecure   bool
    CACert        []byte
    Impersonation *rest.ImpersonationConfig

    // cluster initialization options
    DisableCRDWatch bool
    CacheConfig     *ristretto.Config[string, any]
}

// RegisterOption is the registration-time only option.
type RegisterOption func(*RegisterParams)

// RegisterProxyURL sets HTTP proxy URL for client requests.
func RegisterProxyURL(u string) RegisterOption {
    return func(p *RegisterParams) { p.ProxyURL = u }
}

// RegisterProxyFunc sets custom proxy function.
func RegisterProxyFunc(fn func(*http.Request) (*url.URL, error)) RegisterOption {
    return func(p *RegisterParams) { p.ProxyFunc = fn }
}

// RegisterTimeout sets request timeout on rest.Config.
func RegisterTimeout(d time.Duration) RegisterOption {
    return func(p *RegisterParams) { p.Timeout = d }
}

// RegisterQPS sets QPS for rest.Config.
func RegisterQPS(q float32) RegisterOption {
    return func(p *RegisterParams) { p.QPS = &q }
}

// RegisterBurst sets Burst for rest.Config.
func RegisterBurst(b int) RegisterOption {
    return func(p *RegisterParams) { p.Burst = &b }
}

// RegisterUserAgent sets custom user-agent for rest.Config.
func RegisterUserAgent(ua string) RegisterOption {
    return func(p *RegisterParams) { p.UserAgent = ua }
}

// RegisterTLSInsecure enables insecure TLS on rest.Config.
func RegisterTLSInsecure() RegisterOption {
    return func(p *RegisterParams) { p.TLSInsecure = true }
}

// RegisterCACert sets CAData for TLS verification.
func RegisterCACert(pem []byte) RegisterOption {
    return func(p *RegisterParams) { p.CACert = pem }
}

// RegisterImpersonation sets impersonation config on rest.Config.
func RegisterImpersonation(user string, groups []string, extra map[string][]string) RegisterOption {
    return func(p *RegisterParams) {
        p.Impersonation = &rest.ImpersonationConfig{
            UserName: user,
            Groups:   groups,
            Extra:    extra,
        }
    }
}

// RegisterDisableCRDWatch disables CRD watching during initialization.
func RegisterDisableCRDWatch() RegisterOption {
    return func(p *RegisterParams) { p.DisableCRDWatch = true }
}

// RegisterCacheConfig sets custom cache configuration for the cluster.
func RegisterCacheConfig(cfg *ristretto.Config[string, any]) RegisterOption {
    return func(p *RegisterParams) { p.CacheConfig = cfg }
}