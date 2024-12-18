package starlibs

import (
	"encoding/json"
	"fmt"
	urlpkg "net/url"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/qri-io/starlib/util"
	"github.com/wzshiming/socks"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

type sRestyModule struct {
	client *req.Client
}

func restyModule() *starlarkstruct.Struct {
	var httpClient = req.NewClient().
		EnableAllowGetMethodPayload().
		SetLogger(logx.GetSubLogger())
	var r = &sRestyModule{
		client: httpClient,
	}

	return r.Struct()
}

func (r *sRestyModule) Struct() *starlarkstruct.Struct {
	return starlarkstruct.FromStringDict(starlarkstruct.Default, r.StringDict())
}

func (r *sRestyModule) StringDict() starlark.StringDict {
	return starlark.StringDict{
		"enable_debug_log": starlark.NewBuiltin("enable_debug_log", r.enableOrDisable(func() error {
			r.client.EnableDebugLog()
			return nil
		})),
		"disable_debug_log": starlark.NewBuiltin("disable_debug_log", r.enableOrDisable(func() error {
			r.client.DisableDebugLog()
			return nil
		})),
		"enable_trace_all": starlark.NewBuiltin("enable_trace_all", r.enableOrDisable(func() error {
			r.client.EnableTraceAll()
			return nil
		})),
		"disable_trace_all": starlark.NewBuiltin("enable_trace_all", r.enableOrDisable(func() error {
			r.client.DisableTraceAll()
			return nil
		})),
		"enable_dump_all": starlark.NewBuiltin("enable_dump_all", r.enableOrDisable(func() error {
			r.client.EnableDumpAll()
			return nil
		})),
		"disable_dump_all": starlark.NewBuiltin("disable_dump_all", r.enableOrDisable(func() error {
			r.client.DisableDumpAll()
			return nil
		})),
		"enable_h2c": starlark.NewBuiltin("enable_h2c", r.enableOrDisable(func() error {
			r.client.EnableH2C()
			return nil
		})),
		"disable_h2c": starlark.NewBuiltin("disable_h2c", r.enableOrDisable(func() error {
			r.client.DisableH2C()
			return nil
		})),
		"enable_http3": starlark.NewBuiltin("enable_http3", r.enableOrDisable(func() error {
			r.client.EnableHTTP3()
			return nil
		})),
		"set_base_url":            starlark.NewBuiltin("set_base_url", r.setBaseUrl),
		"set_proxy_url":           starlark.NewBuiltin("set_proxy_url", r.setProxyUrl),
		"set_user_agent":          starlark.NewBuiltin("set_user_agent", r.setUserAgent),
		"set_http_fingerprint":    starlark.NewBuiltin("set_http_fingerprint", r.setHttpFingerprint),
		"set_tls_fingerprint":     starlark.NewBuiltin("set_tls_fingerprint", r.setTlsFingerprint),
		"set_common_retry_count":  starlark.NewBuiltin("set_common_retry_count", r.setRetryCount),
		"set_common_headers":      starlark.NewBuiltin("set_common_headers", r.setHeaders),
		"set_common_path_params":  starlark.NewBuiltin("set_common_path_params", r.setPathParams),
		"set_common_basic_auth":   starlark.NewBuiltin("set_common_basic_auth", r.setBasicAuth),
		"set_common_bearer_auth":  starlark.NewBuiltin("set_common_bearer_auth", r.setBearerAuth),
		"set_common_digest_auth":  starlark.NewBuiltin("set_common_digest_auth", r.setDigestAuth),
		"set_common_content_type": starlark.NewBuiltin("set_common_content_type", r.setContentType),
		"set_common_query_params": starlark.NewBuiltin("set_common_query_params", r.setQueryParams),
		"set_common_form_data":    starlark.NewBuiltin("set_common_form_data", r.setFormData),
		"add_common_query_params": starlark.NewBuiltin("add_common_query_params", r.addQueryParams),
		"get":                     starlark.NewBuiltin("get", r.reqMethod("get")),
		"post":                    starlark.NewBuiltin("post", r.reqMethod("post")),
		"head":                    starlark.NewBuiltin("head", r.reqMethod("head")),
		"delete":                  starlark.NewBuiltin("delete", r.reqMethod("delete")),
		"put":                     starlark.NewBuiltin("put", r.reqMethod("put")),
		"patch":                   starlark.NewBuiltin("patch", r.reqMethod("patch")),
		"options":                 starlark.NewBuiltin("options", r.reqMethod("options")),
		"new_req":                 starlark.NewBuiltin("new_req", r.newReq),
	}
}

func (r *sRestyModule) enableOrDisable(fn func() error) func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := fn(); err != nil {
			return nil, fmt.Errorf("enable_or_diable_error: %v", err)
		}
		return r.Struct(), nil
	}
}

func (r *sRestyModule) setBaseUrl(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var baseUrl string
	if err := starlark.UnpackArgs("set_base_url", args, kwargs, "base_url", &baseUrl); err != nil {
		return nil, err
	}
	r.client.SetBaseURL(baseUrl)
	return r.Struct(), nil
}

func (r *sRestyModule) setProxyUrl(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var proxyUrl string
	if err := starlark.UnpackArgs("set_proxy_url", args, kwargs, "proxy_url", &proxyUrl); err != nil {
		return nil, err
	}
	if proxyUrl == "" {
		return nil, nil
	}
	// 判断是否为sock代理
	u, err := urlpkg.Parse(proxyUrl)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "http":
		r.client.SetProxyURL(proxyUrl)
	default:
		dial, err := socks.NewDialer(proxyUrl)
		if err != nil {
			return nil, err
		}
		r.client.SetDial(dial.DialContext)
	}

	return r.Struct(), nil
}

func (r *sRestyModule) setRetryCount(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var retryCount int
	if err := starlark.UnpackArgs("set_retry_count", args, kwargs, "retry_count", &retryCount); err != nil {
		return nil, err
	}
	r.client.SetCommonRetryCount(retryCount)
	return r.Struct(), nil
}

func (r *sRestyModule) setUserAgent(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var userAgent string
	if err := starlark.UnpackArgs("set_user_agent", args, kwargs, "user_agent", &userAgent); err != nil {
		return nil, err
	}
	r.client.SetUserAgent(userAgent)
	return r.Struct(), nil
}

func (r *sRestyModule) setHeaders(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var headers = &starlark.Dict{}
	if err := starlark.UnpackArgs("set_headers", args, kwargs, "headers", &headers); err != nil {
		return nil, err
	}
	_map, err := convertDictToMap(headers)
	if err != nil {
		return nil, err
	}
	r.client.SetCommonHeaders(_map)
	return r.Struct(), nil
}

func (r *sRestyModule) setPathParams(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pathParams = &starlark.Dict{}
	if err := starlark.UnpackArgs("set_path_params", args, kwargs, "path_params", &pathParams); err != nil {
		return nil, err
	}
	_map, err := convertDictToMap(pathParams)
	if err != nil {
		return nil, err
	}
	r.client.SetCommonPathParams(_map)
	return r.Struct(), nil
}

func (r *sRestyModule) setHttpFingerprint(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var fingerprint string
	if err := starlark.UnpackArgs("set_http_fingerprint", args, kwargs, "fingerprint?", &fingerprint); err != nil {
		return nil, err
	}
	switch fingerprint {
	case "chrome", "Chrome":
		r.client.ImpersonateChrome()
	case "firefox", "Firefox":
		r.client.ImpersonateFirefox()
	case "safari", "Safari":
		r.client.ImpersonateSafari()
	default:
		r.client.ImpersonateChrome()
	}
	return r.Struct(), nil
}

func (r *sRestyModule) setTlsFingerprint(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var fingerprint string
	if err := starlark.UnpackArgs("set_tls_fingerprint", args, kwargs, "fingerprint?", &fingerprint); err != nil {
		return nil, err
	}
	switch fingerprint {
	case "chrome", "Chrome":
		r.client.SetTLSFingerprintChrome()
	case "firefox", "Firefox":
		r.client.SetTLSFingerprintFirefox()
	case "edge", "Edge":
		r.client.SetTLSFingerprintEdge()
	case "qq", "QQ":
		r.client.SetTLSFingerprintQQ()
	case "safari", "Safari":
		r.client.SetTLSFingerprintSafari()
	case "360":
		r.client.SetTLSFingerprint360()
	case "ios", "IOS":
		r.client.SetTLSFingerprintIOS()
	case "android", "Android":
		r.client.SetTLSFingerprintAndroid()
	default:
		r.client.SetTLSFingerprintRandomized()
	}
	return r.Struct(), nil
}

func (r *sRestyModule) setBasicAuth(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var username, password string
	if err := starlark.UnpackArgs("set_common_basic_auth", args, kwargs, "username", &username, "password", &password); err != nil {
		return nil, err
	}
	r.client.SetCommonBasicAuth(username, password)
	return r.Struct(), nil
}

func (r *sRestyModule) setBearerAuth(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var token string
	if err := starlark.UnpackArgs("set_common_bearer_auth", args, kwargs, "token", &token); err != nil {
		return nil, err
	}
	r.client.SetCommonBearerAuthToken(token)
	return r.Struct(), nil
}

func (r *sRestyModule) setDigestAuth(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var username, password string
	if err := starlark.UnpackArgs("set_common_digest_auth", args, kwargs, "username", &username, "password", &password); err != nil {
		return nil, err
	}
	r.client.SetCommonDigestAuth(username, password)
	return r.Struct(), nil
}

func (r *sRestyModule) setContentType(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var contentType string
	if err := starlark.UnpackArgs("set_common_content_type", args, kwargs, "content_type", &contentType); err != nil {
		return nil, err
	}
	r.client.SetCommonContentType(contentType)
	return r.Struct(), nil
}

func (r *sRestyModule) setQueryParams(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var queryParams = &starlark.Dict{}
	if err := starlark.UnpackArgs("set_common_query_params", args, kwargs, "query_params", &queryParams); err != nil {
		return nil, err
	}
	_map, err := convertDictToMap(queryParams)
	if err != nil {
		return nil, err
	}
	r.client.SetCommonQueryParams(_map)
	return r.Struct(), nil
}

func (r *sRestyModule) setFormData(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var formData = &starlark.Dict{}
	if err := starlark.UnpackArgs("set_common_form_data", args, kwargs, "form_data", &formData); err != nil {
		return nil, err
	}
	_map, err := convertDictToMap(formData)
	if err != nil {
		return nil, err
	}
	r.client.SetCommonFormData(_map)
	return r.Struct(), nil
}

func (r *sRestyModule) addQueryParams(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key starlark.String
	var values starlark.Value
	if err := starlark.UnpackArgs("add_common_query_params", args, kwargs, "key", &key, "values", &values); err != nil {
		return nil, err
	}
	vals, err := asStringSlice(values)
	if err != nil {
		return nil, err
	}
	r.client.AddCommonQueryParams(key.String(), vals...)
	return r.Struct(), nil
}

func (r *sRestyModule) reqMethod(method string) func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	rr := &sRequest{
		client: r.client.Clone().NewRequest(),
	}
	return rr.reqMethod(method)
}

func (r *sRestyModule) newReq(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	rr := &sRequest{
		client: r.client.Clone().NewRequest(),
	}
	return rr.Struct(), nil
}

type sRequest struct {
	client *req.Request
}

func (r *sRequest) Struct() *starlarkstruct.Struct {
	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"set_retry_count":  starlark.NewBuiltin("set_retry_count", r.setRetryCount),
		"set_headers":      starlark.NewBuiltin("set_headers", r.setHeaders),
		"set_path_params":  starlark.NewBuiltin("set_path_params", r.setPathParams),
		"set_basic_auth":   starlark.NewBuiltin("set_basic_auth", r.setBasicAuth),
		"set_bearer_auth":  starlark.NewBuiltin("set_bearer_auth", r.setBearerAuth),
		"set_digest_auth":  starlark.NewBuiltin("set_digest_auth", r.setDigestAuth),
		"set_content_type": starlark.NewBuiltin("set_content_type", r.setContentType),
		"set_query_params": starlark.NewBuiltin("set_query_params", r.setQueryParams),
		"add_query_params": starlark.NewBuiltin("add_query_params", r.addQueryParams),
		"set_body":         starlark.NewBuiltin("set_body", r.setBody),
		"set_body_bytes":   starlark.NewBuiltin("set_body_bytes", r.setBodyBytes),
		"set_form_data":    starlark.NewBuiltin("set_form_data", r.setFormData),
		"set_files":        starlark.NewBuiltin("set_files", r.setFiles),
		"set_output":       starlark.NewBuiltin("set_output", r.setOutput),
		"get":              starlark.NewBuiltin("get", r.reqMethod("get")),
		"post":             starlark.NewBuiltin("post", r.reqMethod("post")),
		"head":             starlark.NewBuiltin("head", r.reqMethod("head")),
		"delete":           starlark.NewBuiltin("delete", r.reqMethod("delete")),
		"put":              starlark.NewBuiltin("put", r.reqMethod("put")),
		"patch":            starlark.NewBuiltin("patch", r.reqMethod("patch")),
		"options":          starlark.NewBuiltin("options", r.reqMethod("options")),
	})
}

func (r *sRequest) setRetryCount(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var retryCount int
	if err := starlark.UnpackArgs("set_retry_count", args, kwargs, "retry_count", &retryCount); err != nil {
		return nil, err
	}
	r.client.SetRetryCount(retryCount)
	return r.Struct(), nil
}

func (r *sRequest) setHeaders(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var headers = &starlark.Dict{}
	if err := starlark.UnpackArgs("set_headers", args, kwargs, "headers", &headers); err != nil {
		return nil, err
	}
	_map, err := convertDictToMap(headers)
	if err != nil {
		return nil, err
	}
	r.client.SetHeaders(_map)
	return r.Struct(), nil
}

func (r *sRequest) setPathParams(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pathParams = &starlark.Dict{}
	if err := starlark.UnpackArgs("set_path_params", args, kwargs, "path_params", &pathParams); err != nil {
		return nil, err
	}
	_map, err := convertDictToMap(pathParams)
	if err != nil {
		return nil, err
	}
	r.client.SetPathParams(_map)
	return r.Struct(), nil
}

func (r *sRequest) setBasicAuth(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var username, password string
	if err := starlark.UnpackArgs("set_basic_auth", args, kwargs, "username", &username, "password", &password); err != nil {
		return nil, err
	}
	r.client.SetBasicAuth(username, password)
	return r.Struct(), nil
}

func (r *sRequest) setBearerAuth(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var token string
	if err := starlark.UnpackArgs("set_bearer_auth", args, kwargs, "token", &token); err != nil {
		return nil, err
	}
	r.client.SetBearerAuthToken(token)
	return r.Struct(), nil
}

func (r *sRequest) setDigestAuth(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var username, password string
	if err := starlark.UnpackArgs("set_digest_auth", args, kwargs, "username", &username, "password", &password); err != nil {
		return nil, err
	}
	r.client.SetDigestAuth(username, password)
	return r.Struct(), nil
}

func (r *sRequest) setContentType(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var contentType string
	if err := starlark.UnpackArgs("set_content_type", args, kwargs, "content_type", &contentType); err != nil {
		return nil, err
	}
	r.client.SetContentType(contentType)
	return r.Struct(), nil
}

func (r *sRequest) setQueryParams(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var queryParams = &starlark.Dict{}
	if err := starlark.UnpackArgs("set_query_params", args, kwargs, "query_params", &queryParams); err != nil {
		return nil, err
	}
	_map, err := convertDictToMap(queryParams)
	if err != nil {
		return nil, err
	}
	r.client.SetQueryParams(_map)
	return r.Struct(), nil
}

func (r *sRequest) setFormData(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var formData = &starlark.Dict{}
	if err := starlark.UnpackArgs("set_form_data", args, kwargs, "form_data", &formData); err != nil {
		return nil, err
	}
	_map, err := convertDictToMap(formData)
	if err != nil {
		return nil, err
	}
	r.client.SetFormData(_map)
	return r.Struct(), nil
}

func (r *sRequest) setFiles(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var files = &starlark.Dict{}
	if err := starlark.UnpackArgs("set_files", args, kwargs, "files", &files); err != nil {
		return nil, err
	}
	_map, err := convertDictToMap(files)
	if err != nil {
		return nil, err
	}
	r.client.SetFiles(_map)
	return r.Struct(), nil
}

func (r *sRequest) setOutput(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var output string
	if err := starlark.UnpackArgs("set_output", args, kwargs, "output", &output); err != nil {
		return nil, err
	}
	r.client.SetOutputFile(output)
	return r.Struct(), nil
}

func (r *sRequest) setBody(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var body starlark.Value
	if err := starlark.UnpackArgs("set_body", args, kwargs, "body", &body); err != nil {
		return nil, err
	}
	// 断言 body 是否为字符串或者字典
	switch body.(type) {
	case starlark.String:
		r.client.SetBodyString(body.(starlark.String).String())
	case *starlark.Dict:
		_map, err := convertDictToMap(body.(*starlark.Dict))
		if err != nil {
			return nil, err
		}
		r.client.SetBodyJsonMarshal(_map)
	default:
		return nil, fmt.Errorf("expected param body to be a string or a dict. got: '%s'", body.Type())
	}
	return r.Struct(), nil
}

func (r *sRequest) setBodyBytes(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var body starlark.Bytes
	if err := starlark.UnpackArgs("set_body_bytes", args, kwargs, "body", &body); err != nil {
		return nil, err
	}
	r.client.SetBodyString(body.String())
	return r.Struct(), nil
}

func (r *sRequest) addQueryParams(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key starlark.String
	var values starlark.Value
	if err := starlark.UnpackArgs("add_query_params", args, kwargs, "key", &key, "values", &values); err != nil {
		return nil, err
	}
	vals, err := asStringSlice(values)
	if err != nil {
		return nil, err
	}
	r.client.AddQueryParams(key.String(), vals...)
	return r.Struct(), nil
}

func (r *sRequest) reqMethod(method string) func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var urlv starlark.String
		var data starlark.Value
		var headers = &starlark.Dict{}
		var pathParams = &starlark.Dict{}
		var queryParams = &starlark.Dict{}
		if err := starlark.UnpackArgs(method, args, kwargs, "url", &urlv, "data?", &data, "headers?", &headers, "path_params?", &pathParams, "query_params?", &queryParams); err != nil {
			return nil, err
		}
		if data != nil {
			_, err := r.setBody(thread, nil, starlark.Tuple{data}, nil)
			if err != nil {
				return nil, err
			}
		}
		if headers != nil {
			_, err := r.setHeaders(thread, nil, starlark.Tuple{headers}, nil)
			if err != nil {
				return nil, err
			}
		}
		if pathParams != nil {
			_, err := r.setPathParams(thread, nil, starlark.Tuple{pathParams}, nil)
			if err != nil {
				return nil, err
			}
		}
		if queryParams != nil {
			_, err := r.setQueryParams(thread, nil, starlark.Tuple{queryParams}, nil)
			if err != nil {
				return nil, err
			}
		}
		rawurl, err := asString(urlv)
		if err != nil {
			return nil, err
		}

		resp, err := r.client.Send(strings.ToUpper(method), rawurl)
		if err != nil {
			return nil, err
		}
		rr := &sResponse{*resp}
		return rr.Struct(), nil
	}
}

// sResponse represents an req response, wrapping a go req.Response with
// starlark methods
type sResponse struct {
	req.Response
}

func (r *sResponse) Struct() *starlarkstruct.Struct {

	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"url":         starlark.String(r.Request.URL.String()),
		"status_code": starlark.MakeInt(r.StatusCode),
		"status":      starlark.String(r.Status),
		"proto":       starlark.String(r.Proto),
		"proto_major": starlark.MakeInt(r.ProtoMajor),
		"proto_minor": starlark.MakeInt(r.ProtoMinor),
		"headers":     r.headersDict(),
		"body":        starlark.NewBuiltin("body", r.text),
		"json":        starlark.NewBuiltin("json", r.json),
	})
}

func (r *sResponse) headersDict() *starlark.Dict {
	d := starlark.NewDict(len(r.Header))
	for k, v := range r.Header {
		_ = d.SetKey(starlark.String(k), starlark.String(strings.Join(v, ", ")))
	}
	return d
}

func (r *sResponse) text(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.String(r.String()), nil
}

func (r *sResponse) json(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var data interface{}
	if err := json.Unmarshal(r.Bytes(), &data); err != nil {
		return nil, err
	}
	return util.Marshal(data)
}
