package ezcomm

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"gitee.com/bon-ami/eztools/v6"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

const (
	tstDefHTTPProt    = "tcp"
	tstDefHTTPMeth    = http.MethodGet
	tstDefHTTPCode    = http.StatusOK
	tstDefHTTPFail    = http.StatusInternalServerError
	tstDefHTTPBody    = EzcName
	tstDefHTTPEcho    = "Echo "
	tstDefHTTPUrlPref = "http://"
)

func tstHTTPSvrShutdown(t *testing.T, svr *HTTPSvr) error {
	if *tstVerbose > 1 {
		t.Log("shutting down server")
	}
	return svr.Shutdown(time.Second * time.Duration(*TstClntNo))
}

// TestHTTPSvr uses TstProt, TstLcl, TstTimeout and TstClntNo
func TestHTTPSvrFS(t *testing.T) {
	Init4Tests(t)
	if !strings.HasPrefix(*TstProt, "tcp") {
		*TstProt = tstDefHTTPProt
	}
	lstnr, err := ListenTCP(t.Log, nil, *TstProt, *TstLcl, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if *tstVerbose > 0 {
		t.Log(lstnr.Addr().String())
	}
	svr := MakeHTTPSvr()
	svr.FS("", *TstRoot, nil)
	ch := svr.Serve(lstnr)
	if *tstVerbose > 1 {
		t.Log("wait for server for", TstTO)
	}
	chClnt := make(chan error, 1)
	defer close(chClnt)
	go func() {
		resp, err := eztools.HTTPSend(tstDefHTTPMeth,
			tstDefHTTPUrlPref+lstnr.Addr().String(),
			"", nil)
		if err != nil {
			//t.Error(err)
			chClnt <- errors.Wrap(err, "Send")
			return
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			//t.Error(err)
			chClnt <- errors.Wrap(err, "Read Body")
			return
		}
		if *tstVerbose > 0 {
			t.Log(string(data))
		}
		chClnt <- nil
	}()
	select {
	case <-time.After(TstTO):
		err = tstHTTPSvrShutdown(t, svr)
	case err = <-chClnt:
		if err != nil {
			t.Error(err)
		}
		err = tstHTTPSvrShutdown(t, svr)
	case err = <-ch:
	}
	if err != nil {
		t.Fatal(err)
	}
	/*if n := runtime.NumGoroutine(); n > 2 {
		<-time.After(time.Second * 5)
		if n := runtime.NumGoroutine(); n > 2 {
			t.Error("routines left (should be 0)", n-2)
		}
	}*/
}

// TestHTTPSvrPst uses TstProt, TstLcl, TstTimeout and TstClntNo
func TestHTTPSvrPst(t *testing.T) {
	Init4Tests(t)
	if !strings.HasPrefix(*TstProt, "tcp") {
		*TstProt = tstDefHTTPProt
	}
	lstnr, err := ListenTCP(t.Log, nil, *TstProt, *TstLcl, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if *tstVerbose > 0 {
		t.Log(lstnr.Addr().String())
	}
	chClnt := make(chan error, 1)
	defer close(chClnt)
	svr := MakeHTTPSvr()
	const (
		bodyInput      = "info"
		bodyInputCtrl  = "input"
		bodyInputName  = "name"
		bodyInputValue = "value"
		bodyPref       = `<form action="/" method="POST">
<` + bodyInputCtrl + " '" + bodyInputName + "' = '" + bodyInput + "' " + bodyInputValue + "='"
		bodySuf = `'>
<p><input type="submit" value="` + `Submit` + `" /></p>
</form>`
	)
	svr.GET("", func(ip string, req *http.Request,
		_ func(string) string) (int, HTTPSvrBody, map[string]string) {
		return tstDefHTTPCode, HTTPSvrBody{
			Tp:  HTTPSvrBodyHTML,
			Str: bodyPref + TstHalo + bodySuf,
		}, nil
	})
	svr.POST("", func(ip string, req *http.Request,
		fun func(string) string) (int, HTTPSvrBody, map[string]string) {
		tstInput := fun(bodyInput)
		if tstInput != TstBye {
			t.Error(tstInput, "got instead of", TstBye)
		}
		//defer func() { chClnt <- nil }()
		return tstDefHTTPCode, HTTPSvrBody{
			Tp:  HTTPSvrBodyHTML,
			Str: bodyPref + bodySuf,
		}, nil
	})
	ch := svr.Serve(lstnr)
	if *tstVerbose > 1 {
		t.Log("wait for server for", TstTO)
	}
	sndPost := func(fld string) {
		go func() {
			uv := url.Values{}
			uv.Set(fld, TstBye)
			ue := uv.Encode()
			resp, err := eztools.HTTPSend(http.MethodPost,
				tstDefHTTPUrlPref+lstnr.Addr().String(),
				eztools.BodyTypeForm, strings.NewReader(ue))
			if err != nil {
				//t.Error(err)
				chClnt <- errors.Wrap(err, "Send Post")
				return
			}
			_, _, data, code, err := eztools.HTTPParseBody(resp, "", nil, nil)
			// check body
			if err != nil {
				//t.Error(err)
				chClnt <- errors.Wrap(err, "Read Body from Response")
				return
			}
			if *tstVerbose > 0 {
				t.Log("body from server:", string(data))
			}
			if code != http.StatusOK {
				//t.Error(code)
				chClnt <- eztools.ErrAccess
				return
			}
			// TODO: parse body html
			chClnt <- nil
		}()
	}
	go func() { //tstHTTPClnt(http.MethodGet, lstnr.Addr().String(), chClnt,
		//nil, nil, func(resp *http.Response) {
		resp, err := eztools.HTTPSend(http.MethodGet,
			tstDefHTTPUrlPref+lstnr.Addr().String(),
			"", nil)
		if err != nil {
			//t.Error(err)
			chClnt <- errors.Wrap(err, "Send Get")
			return
		}
		defer resp.Body.Close()
		var fld string
		var parser func(n *html.Node)
		parser = func(n *html.Node) {
			if n.Type == html.ElementNode &&
				n.Data == bodyInputCtrl {
				for _, a := range n.Attr {
					if a.Key == bodyInputName {
						fld = a.Val
						return
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				parser(c)
			}
		}
		bodyType, _, data, code, err := eztools.HTTPParseBody(resp, "", parser, nil)
		// check body
		if err != nil {
			//t.Error(err)
			chClnt <- errors.Wrap(err, "Read Body from Response")
			return
		}
		if bodyType != eztools.BodyTypeHTML || len(fld) < 1 {
			chClnt <- errors.Wrap(err, "Read Body from Response")
			return
		}
		if *tstVerbose > 0 {
			t.Log("body from server:", string(data))
		}
		if code != http.StatusOK {
			//t.Error(code)
			chClnt <- eztools.ErrAccess
			return
		}
		sndPost(fld)
	}()
	select {
	case <-time.After(TstTO):
		err = tstHTTPSvrShutdown(t, svr)
	case err = <-chClnt:
		if err != nil {
			t.Error(err)
		}
		err = tstHTTPSvrShutdown(t, svr)
	case err = <-ch:
	}
	/*if n := runtime.NumGoroutine(); n > 2 {
		<-time.After(time.Second * 5)
		if n := runtime.NumGoroutine(); n > 2 {
			t.Error("routines left (should be 0)", n-2)
		}
	}*/
}

// TestHTTPSvrHdr uses TstProt, TstLcl, TstTimeout and TstClntNo
func TestHTTPSvrHdr(t *testing.T) {
	Init4Tests(t)
	if !strings.HasPrefix(*TstProt, "tcp") {
		*TstProt = tstDefHTTPProt
	}
	lstnr, err := ListenTCP(t.Log, nil, *TstProt, *TstLcl, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if *tstVerbose > 0 {
		t.Log(lstnr.Addr().String())
	}
	hdrs := make(map[string]string)
	for i := '1'; i < '5'; i++ {
		hdrs["Header"+string(i)] = "Value" + string(i)
	}
	chClnt := make(chan error, 1)
	defer close(chClnt)
	svr := MakeHTTPSvr()
	svr.GET("", func(peer string, req *http.Request, _ func(string) string) (
		int, HTTPSvrBody, map[string]string) {
		data, err := io.ReadAll(req.Body)
		//bd, err := req.GetBody()
		if err != nil {
			//t.Error(err)
			chClnt <- errors.Wrap(err, "Read Body from Request")
			return tstDefHTTPFail, HTTPSvrBody{}, nil
		}
		defer req.Body.Close()
		/*bdBuf := make([]byte, len(tstDefHTTPBody))
		if _, err = bd.Read(bdBuf); err != nil {
			t.Error(err)
			return
		}*/
		retHdrs := make(map[string]string)
		for hdr1 := range hdrs {
			vals, ok := req.Header[hdr1]
			if !ok || len(vals) < 1 {
				t.Error(hdr1, "NOT found in request")
				continue
			}
			for i, v1 := range vals {
				if i > 0 {
					retHdrs[hdr1] += ", "
				} else {
					retHdrs[hdr1] = tstDefHTTPEcho
				}
				retHdrs[hdr1] += v1
			}
		}
		if *tstVerbose > 2 {
			t.Log("headers to echo:", retHdrs)
		}
		return tstDefHTTPCode, HTTPSvrBody{
			Tp:  HTTPSvrBodyString,
			Str: tstDefHTTPEcho + string(data)}, retHdrs
	})
	ch := svr.Serve(lstnr)
	if *tstVerbose > 1 {
		t.Log("wait for server for", TstTO)
	}
	go func() { //tstHTTPClnt("", lstnr.Addr().String(), chClnt,
		/*hdrs, strings.NewReader(tstDefHTTPBody),
		func(resp *http.Response) {*/
		resp, err := eztools.HTTPSendHdr(tstDefHTTPMeth,
			tstDefHTTPUrlPref+lstnr.Addr().String(),
			eztools.BodyTypeText, strings.NewReader(tstDefHTTPBody), hdrs)
		if err != nil {
			//t.Error(err)
			chClnt <- errors.Wrap(err, "Send")
			return
		}
		defer resp.Body.Close()
		// check headers
		if *tstVerbose > 2 {
			t.Log("headers to match:", resp.Header)
		}
		for hdr1, val1 := range hdrs {
			vals, ok := resp.Header[hdr1]
			if !ok || len(vals) < 1 {
				t.Error(val1, "NOT found in response")
				continue
			}
			val := strings.Join(vals, "")
			if val != tstDefHTTPEcho+val1 {
				t.Error(val, "!=", tstDefHTTPEcho+val1,
					"in response")
			}
		}
		// check body
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			//t.Error(err)
			chClnt <- errors.Wrap(err, "Read Body from Response")
			return
		}
		if *tstVerbose > 0 {
			t.Log("body from server:", string(data))
		}
		if string(data) != tstDefHTTPEcho+tstDefHTTPBody {
			t.Error(string(data), "!=",
				tstDefHTTPEcho+tstDefHTTPBody,
				"in body")
		}
		chClnt <- nil
	}()
	select {
	case <-time.After(TstTO):
		err = tstHTTPSvrShutdown(t, svr)
	case err = <-chClnt:
		if err != nil {
			t.Error(err)
		}
		err = tstHTTPSvrShutdown(t, svr)
	case err = <-ch:
	}
	/*if n := runtime.NumGoroutine(); n > 2 {
		<-time.After(time.Second * 5)
		if n := runtime.NumGoroutine(); n > 2 {
			t.Error("routines left (should be 0)", n-2)
		}
	}*/
}
