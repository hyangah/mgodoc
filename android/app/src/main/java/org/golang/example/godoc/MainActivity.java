package org.golang.example.godoc;

import android.content.Intent;
import android.content.res.TypedArray;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.support.v7.app.ActionBar;
import android.support.v7.app.ActionBarActivity;
import android.util.Log;
import android.view.KeyEvent;
import android.view.Window;
import android.webkit.WebResourceResponse;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import java.io.ByteArrayInputStream;
import java.io.InputStream;

import go.godoc.Godoc;

public class MainActivity extends ActionBarActivity {
    private static final String DEBUG_TAG = "BrowseGodoc";
    private static final String MIME = "text/html";
    private static final String ENCODING = "utf-8";

    private ActionBar mActionBar;
    private WebView mWebView;

    private byte[] loadPage(String url) {
        try {
            Godoc.Response resp = Godoc.Serve(url);
            return resp.getBody();
        } catch (Exception e) {
            return e.toString().getBytes();
        }
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        getWindow().requestFeature(Window.FEATURE_ACTION_BAR_OVERLAY);
        super.onCreate(savedInstanceState);

        setContentView(R.layout.activity_main);
        final TypedArray styledAttributes = getTheme().obtainStyledAttributes(
                new int[]{android.R.attr.actionBarSize});

        styledAttributes.recycle();

        mActionBar = getSupportActionBar();
        mActionBar.hide();

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
            WebView.setWebContentsDebuggingEnabled(true);
        }

        setContentView(R.layout.activity_main);

        mWebView = (WebView) findViewById(R.id.webview);

        WebSettings webSettings = mWebView.getSettings();
        webSettings.setJavaScriptEnabled(true);
        webSettings.setBuiltInZoomControls(true);

        if (savedInstanceState != null) {
            mWebView.restoreState(savedInstanceState);
        } else {
            try {
                String url = "https://golang.org/pkg/";
                Godoc.Response resp = Godoc.Serve(url);
                String htmlStr = new String(resp.getBody(), "UTF-8");
                mWebView.loadDataWithBaseURL(url, htmlStr, MIME, ENCODING, null);
            } catch (Exception e) {
                Log.wtf(DEBUG_TAG, e.toString());
            }
        }

        mWebView.setWebViewClient(new MyWebViewClient());
    }

    protected void onSaveInstanceState(Bundle outState) {
        mWebView.saveState(outState);
    }

    @Override
    public boolean onKeyDown(int keyCode, KeyEvent event) {
        WebView myWebView = (WebView) findViewById(R.id.webview);
        // Check if the key event was the Back button and if there's history.
        if ((keyCode == KeyEvent.KEYCODE_BACK) && myWebView.canGoBack()) {
            myWebView.goBack();
            return true;
        }
        // If it wasn't the Back key or there's no web page history,
        // bubble up to the default system behavior (probably exit the activity).
        return super.onKeyDown(keyCode, event);
    }

    private class MyWebViewClient extends WebViewClient {
        @Override
        public boolean shouldOverrideUrlLoading(WebView view, String url) {
            if (Uri.parse(url).getHost().equals("golang.org")) {
                // This is my web site, so do not override; let my WebView load the page.
                return false;
            }
            // Otherwise, the link is not for a page on my site, so launch another Activity that handles URLs
            Intent intent = new Intent(Intent.ACTION_VIEW, Uri.parse(url));
            startActivity(intent);
            return true;
        }

        @Override
        public WebResourceResponse shouldInterceptRequest(WebView view, String url) {
            // TODO: change to the new API shouldInterceptRequest(WebView view, WebResourceRequest req)
            try {
                Godoc.Response resp = Godoc.Serve(url.toString());
                if (resp.getStatusCode() != 200) {  // OK
                    return null;
                }

                String mime = resp.Header("Content-Type");
                String encoding = "";
                if (mime == null || mime.equals("text/html")) {
                    encoding = "utf-8";
                }
                InputStream stream = new ByteArrayInputStream(resp.getBody());
                if (stream != null) {
                    return new WebResourceResponse(mime, encoding, stream);
                }
            } catch (Exception e) {
                return null;
            }
            return null;
        }
    }
}
