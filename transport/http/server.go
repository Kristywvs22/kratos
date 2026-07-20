@@ -10,6 +10,7 @@
 import (
     "context"
     "net/http"
+    "errors"
 )

 type GreeterHTTPServer interface {
@@ -25,14 +26,20 @@
 func _Greeter_SayHello0_HTTP_Handler(srv GreeterHTTPServer) http.HandlerFunc {
     return func(w http.ResponseWriter, r *http.Request) {
         ctx := r.Context()
-        var in HelloRequest
-        // 1. Decoding happens BEFORE the middleware chain is executed
-        if err := ctx.Bind(&in); err != nil {
-            http.Error(w, "Bad Request", http.StatusBadRequest)
-            return
-        }
-        http.SetOperation(ctx, OperationGreeterSayHello)
-        h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
-            return srv.SayHello(ctx, req.(*HelloRequest))
-        })
-        out, err := h(ctx, &in)
-        if err != nil {
-            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
-            return
-        }
+        http.SetOperation(ctx, OperationGreeterSayHello)
+        h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
+            var in HelloRequest
+            if err := ctx.Bind(&in); err != nil {
+                return nil, errors.Wrap(err, "failed to bind request")
+            }
+            return srv.SayHello(ctx, &in)
+        })
+        out, err := h(ctx, nil)
+        if err != nil {
+            if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
+                http.Error(w, "Request Timeout", http.StatusRequestTimeout)
+                return
+            }
+            http.Error(w, "Bad Request", http.StatusBadRequest)
+            return
+        }
         w.WriteHeader(http.StatusOK)
         if err := json.NewEncoder(w).Encode(out); err != nil {
             http.Error(w, "Internal Server Error", http.StatusInternalServerError)
```

### Explanation

1. **Move Binding Inside Middleware Chain**: The `ctx.Bind(&in)` call is moved inside the middleware chain. This ensures that the middleware chain is executed even if the request fails to bind.
2. **Error Handling**: If the binding fails, an error is returned from the middleware function, and the appropriate HTTP status code (`400 Bad Request`) is set.
3. **Middleware Chain Execution**: The middleware chain is now guaranteed to execute, allowing logging, tracing, and other observability middlewares to capture the request metadata and any decoding errors.

This patch should meet the acceptance criteria and ensure that the middleware chain is executed for all incoming requests, including those with malformed JSON bodies.