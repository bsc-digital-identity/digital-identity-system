package zkprequest

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

type Handler struct {
	svc *Service
	log *slog.Logger
}

// Domyślny logger (text na stdout)
func NewHandler(s *Service) *Handler {
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return &Handler{svc: s, log: l}
}

// ---------- Helpers ----------

// Zwraca: humanURL (HTML), descriptorURL (JSON dla walleta), deeplink (opcjonalny)
func makeWalletLink(audience, requestID string) (humanURL, descriptorURL, deeplink string) {
	humanURL = fmt.Sprintf("%s/v1/presentations/%s", audience, requestID)
	descriptorURL = fmt.Sprintf("%s/v1/presentations/%s/descriptor", audience, requestID)
	deeplink = fmt.Sprintf("zkwallet://present?request_uri=%s", url.QueryEscape(descriptorURL))
	return
}

func qrBase64(data string) string {
	png, _ := qrcode.Encode(data, qrcode.Medium, 256)
	return base64.StdEncoding.EncodeToString(png)
}

// ---------- Optional: Gin middleware ----------
func (h *Handler) GinRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		lat := time.Since(start)
		h.log.Info("http_request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"ip", c.ClientIP(),
			"latency_ms", lat.Milliseconds(),
			"user_agent", c.Request.UserAgent(),
			"request_id", c.GetString("request_id"),
		)
	}
}

// ---------- Endpoints ----------

// POST /v1/presentations/create
func (h *Handler) CreatePresentation(c *gin.Context) {
	var in CreatePresentationIn
	if err := c.ShouldBindJSON(&in); err != nil {
		h.log.Warn("create_presentation.bad_json", "error", err.Error(), "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json: " + err.Error()})
		return
	}

	start := time.Now()
	req, err := h.svc.CreateRequestFromSchema(in.SchemaJSON, time.Now())
	if err != nil {
		h.log.Error("create_presentation.create_failed", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if in.CallbackURL != "" {
		if stored, ok := h.svc.Store.Load(req.RequestID); ok {
			stored.CallbackURL = in.CallbackURL
			stored.CallbackSecret = in.CallbackSecret
			_ = h.svc.Store.Save(stored)
			h.log.Info("create_presentation.callback_attached",
				"request_id", req.RequestID,
				"callback_url", in.CallbackURL,
				"has_secret", in.CallbackSecret != "",
			)
		}
	}

	requestURL, descriptorURL, deeplink := makeWalletLink(h.svc.Audience, req.RequestID)
	h.log.Info("create_presentation.ok",
		"request_id", req.RequestID,
		"audience", h.svc.Audience,
		"expires_at_unix", req.ExpiresAt,
		"latency_ms", time.Since(start).Milliseconds(),
	)

	c.JSON(http.StatusOK, CreatePresentationOut{
		Request:     req,
		RequestURL:  requestURL,
		DeepLink:    deeplink,
		QRPngBase64: qrBase64(descriptorURL),
	})
}

// POST /v1/presentations/verify
func (h *Handler) VerifyPresentation(c *gin.Context) {
	var in VerifyIn
	if err := c.ShouldBindJSON(&in); err != nil {
		h.log.Warn("verify_presentation.bad_json", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json: " + err.Error()})
		return
	}

	start := time.Now()
	_, err := h.svc.VerifySubmission(ProofSubmission{
		RequestID:    in.RequestID,
		ZkpBlobB64:   in.ZkpBlobB64,
		PublicInputs: in.PublicInputs,
		Challenge:    in.Challenge,
	})
	if err != nil {
		h.log.Warn("verify_presentation.failed",
			"request_id", in.RequestID,
			"error", err.Error(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": err.Error()})
		return
	}

	h.log.Info("verify_presentation.ok",
		"request_id", in.RequestID,
		"latency_ms", time.Since(start).Milliseconds(),
	)
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"verified_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// GET /v1/presentations/:request_id
func (h *Handler) ShowPresentation(c *gin.Context) {
	requestID := c.Param("request_id")
	req, ok := h.svc.Store.Load(requestID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown request"})
		return
	}
	if time.Now().Unix() > req.ExpiresAt {
		c.JSON(http.StatusGone, gin.H{"error": "expired"})
		return
	}

	humanURL, descriptorURL, deeplink := makeWalletLink(h.svc.Audience, requestID)

	qrURL := "https://quickchart.io/qr?text=" +
		url.QueryEscape(descriptorURL) +
		"&size=320"

	h.log.Info("show_presentation.render",
		"request_id", requestID,
		"audience", h.svc.Audience,
		"deeplink_len", len(deeplink),
		"qr_url", qrURL,
	)

	c.HTML(http.StatusOK, "zkp_request_presentation.html", gin.H{
		"RequestID": requestID,
		"HumanURL":  humanURL,
		"QRURL":     qrURL,
	})
}

// GET /v1/presentations/:request_id/descriptor
// Jawny JSON dla walleta (to samo co negotiation powyżej)
func (h *Handler) Descriptor(c *gin.Context) {
	requestID := c.Param("request_id")
	req, ok := h.svc.Store.Load(requestID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown request"})
		return
	}
	if time.Now().Unix() > req.ExpiresAt {
		c.JSON(http.StatusGone, gin.H{"error": "expired"})
		return
	}

	base := h.svc.Audience

	var challenge string
	if v, ok := req.PublicInputs["challenge"]; ok && v != nil {
		challenge = fmt.Sprint(v)
	}

	// NEW: wyciągamy nonce wygenerowany w CreateRequestFromSchema
	var nonce string
	if v, ok := req.PublicInputs["nonce"]; ok && v != nil {
		nonce = fmt.Sprint(v)
	}

	vkURL := fmt.Sprintf("%s/v1/artifacts/%s/vk", base, req.SchemaHash)
	pkURL := fmt.Sprintf("%s/v1/artifacts/%s/pk", base, req.SchemaHash)
	schemaURI := fmt.Sprintf("%s/v1/schemas/%s", base, req.SchemaHash)
	submitURL := req.ResponseURI

	c.JSON(http.StatusOK, gin.H{
		"request_id": req.RequestID,
		"audience":   base,
		"expires_at": req.ExpiresAt,
		"challenge":  challenge,
		"nonce":      nonce, // <--- TU
		"schema": gin.H{
			"hash": req.SchemaHash,
			"uri":  schemaURI,
		},
		"artifacts": gin.H{
			"vk_url": vkURL,
			"pk_url": pkURL,
		},
		"submit_url": submitURL,
	})
}

// GET /v1/presentations/:request_id/status
func (h *Handler) Status(c *gin.Context) {
	requestID := c.Param("request_id")
	req, ok := h.svc.Store.Load(requestID)
	if !ok {
		h.log.Info("status.unknown_or_consumed", "request_id", requestID)
		c.JSON(http.StatusOK, gin.H{"state": "unknown_or_consumed"})
		return
	}
	if time.Now().Unix() > req.ExpiresAt {
		h.log.Info("status.expired", "request_id", requestID)
		c.JSON(http.StatusOK, gin.H{"state": "expired"})
		return
	}
	h.log.Info("status.pending", "request_id", requestID)
	c.JSON(http.StatusOK, gin.H{"state": "pending"})
}

func (h *Handler) GetVK(c *gin.Context) {
	hash := c.Param("hash")

	h.svc.cacheMu.RLock()
	vk := h.svc.vkCache[hash]
	h.svc.cacheMu.RUnlock()

	if len(vk) == 0 {
		c.String(404, "vk not found")
		return
	}

	c.Data(200, "application/octet-stream", vk)
}

func (h *Handler) GetPK(c *gin.Context) {
	hash := c.Param("hash")

	h.svc.cacheMu.RLock()
	pk := h.svc.pkCache[hash]
	h.svc.cacheMu.RUnlock()

	if len(pk) == 0 {
		c.String(404, "pk not found")
		return
	}

	c.Data(200, "application/octet-stream", pk)
}

// GET /v1/presentations/:request_id/result
func (h *Handler) Result(c *gin.Context) {
	requestID := c.Param("request_id")

	// 🔥 1. Najpierw sprawdź, czy DI już ma finalny verdict (verified/failed/expired)
	if v, ok := h.svc.getVerdict(requestID); ok {
		h.log.Info("result.cached",
			"request_id", requestID,
			"state", v.State,
			"ok", v.OK,
			"reason", v.Reason,
			"verified_at", v.VerifiedAt,
		)

		out := gin.H{
			"state": v.State,
			"ok":    v.OK,
		}
		if v.Reason != "" {
			out["reason"] = v.Reason
		}
		if !v.VerifiedAt.IsZero() {
			out["verified_at"] = v.VerifiedAt.Format(time.RFC3339)
		}

		c.JSON(http.StatusOK, out)
		return
	}

	// 🔥 2. Jeśli verdictu jeszcze nie ma → pending
	req, ok := h.svc.Store.Load(requestID)
	if !ok {
		h.log.Info("result.unknown", "request_id", requestID)
		c.JSON(http.StatusOK, gin.H{"state": "unknown"})
		return
	}

	if time.Now().Unix() > req.ExpiresAt {
		h.log.Info("result.expired", "request_id", requestID)
		c.JSON(http.StatusOK, gin.H{"state": "expired"})
		return
	}

	h.log.Info("result.pending", "request_id", requestID)
	c.JSON(http.StatusOK, gin.H{"state": "pending"})
}

// POST /v1/presentations/verify-blocking
type verifyBlockingIn struct {
	SchemaJSON  string `json:"schema_json" binding:"required"`
	TimeoutSecs int    `json:"timeout_secs,omitempty"`
}

type verifyBlockingOut struct {
	OK        bool   `json:"ok"`
	Reason    string `json:"reason,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func (h *Handler) VerifyBlocking(c *gin.Context) {
	var in verifyBlockingIn
	if err := c.ShouldBindJSON(&in); err != nil {
		h.log.Warn("verify_blocking.bad_json", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json: " + err.Error()})
		return
	}

	start := time.Now()
	req, err := h.svc.CreateRequestFromSchema(in.SchemaJSON, time.Now())
	if err != nil {
		h.log.Error("verify_blocking.create_failed", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeout := time.Duration(in.TimeoutSecs) * time.Second
	if in.TimeoutSecs <= 0 {
		timeout = 3 * time.Minute
	}
	out, _ := h.svc.WaitForResult(req.RequestID, timeout)

	if out.OK {
		h.log.Info("verify_blocking.ok",
			"request_id", req.RequestID,
			"waited_ms", time.Since(start).Milliseconds(),
		)
		c.JSON(http.StatusOK, verifyBlockingOut{OK: true, RequestID: req.RequestID})
		return
	}

	h.log.Info("verify_blocking.failed",
		"request_id", req.RequestID,
		"reason", out.Reason,
		"waited_ms", time.Since(start).Milliseconds(),
	)
	c.JSON(http.StatusUnauthorized, verifyBlockingOut{
		OK:        false,
		Reason:    out.Reason,
		RequestID: req.RequestID,
	})
}

// GET /v1/schemas/:hash
// Zwraca JSON schemy dla danego schema_hash (np. "sha256:...").
func (h *Handler) Schema(c *gin.Context) {
	hash := c.Param("hash")

	h.log.Info("schema.fetch", "hash", hash)

	// wyciągamy schemę z cache serwisu
	h.svc.cacheMu.RLock()
	body, ok := h.svc.schemaCache[hash]
	h.svc.cacheMu.RUnlock()

	if !ok || len(body) == 0 {
		h.log.Warn("schema.not_found", "hash", hash)
		c.String(http.StatusNotFound, "schema not found")
		return
	}

	// oddajemy dokładnie ten canonical JSON, który był użyty do setupu VK
	c.Data(http.StatusOK, "application/json", body)
}
