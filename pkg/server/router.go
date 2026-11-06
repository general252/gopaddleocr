package server

import (
	"log"
	"net/http"

	"github.com/multippt/gopaddleocr/pkg/ocr"
)

type Server struct {
	server         *http.ServeMux
	ocrEngine      *ocr.Engine
	sessionManager *SessionManager
}

func NewServer() *Server {
	return &Server{
		sessionManager: NewSessionManager(),
	}
}

func (s *Server) registerHandlers() {
	s.server.HandleFunc("GET /health", s.handleHealth)
	s.server.HandleFunc("POST /ocr", s.handleOCR)
	s.server.HandleFunc("POST /detect", s.handleDetect)
	s.server.HandleFunc("POST /session/create", s.handleSessionCreate)
	s.server.HandleFunc("POST /session/detect", s.handleSessionDetect)
	s.server.HandleFunc("POST /session/ocr", s.handleSessionOCR)
	s.server.HandleFunc("POST /session/delete", s.handleSessionDelete)

	s.server.HandleFunc("GET /home", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(ocrHTML))
	})
}

func (s *Server) Start(listenAddr string) {
	// Build and preload the OCR engine in the background.
	ocrEngine := ocr.NewEngine()
	go func() {
		if err := ocrEngine.Init(); err != nil {
			log.Printf("engine preload error: %v", err)
		} else {
			log.Println("engine loaded successfully")
		}
	}()
	defer func() {
		_ = ocrEngine.Close()
	}()
	s.ocrEngine = ocrEngine

	s.server = http.NewServeMux()
	s.registerHandlers()

	log.Printf("OCR service listening on http://127.0.0.1%s/home", listenAddr)
	if err := http.ListenAndServe(listenAddr, s.server); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// 内嵌的 HTML 页面（无反引号冲突，用字符串拼接）
const ocrHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>OCR 文字识别</title>
<style>
* { box-sizing: border-box; font-family: system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif; }
body { background: #f5f7fc; margin: 0; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1.5rem; }
.card { max-width: 1300px; width: 100%; background: white; border-radius: 2rem; box-shadow: 0 25px 50px -12px rgba(0,0,0,0.25); overflow: hidden; display: flex; flex-direction: column; }
.header { padding: 1.5rem 2rem; border-bottom: 1px solid #e9eef3; }
.header h1 { margin: 0; font-size: 1.8rem; font-weight: 600; color: #0b2b4a; display: flex; align-items: center; gap: 0.75rem; }
.header h1 span { background: #2563eb; color: white; font-size: 0.9rem; font-weight: 500; padding: 0.2rem 0.8rem; border-radius: 30px; }
.main { display: flex; flex-wrap: wrap; padding: 2rem; gap: 2rem; }
.left { flex: 1 1 380px; display: flex; flex-direction: column; gap: 1.5rem; }
.right { flex: 2 1 600px; display: flex; flex-direction: column; gap: 1rem; min-width: 0; }
.upload-area { background: #f8fafd; border: 2px dashed #cbd5e1; border-radius: 1.5rem; padding: 2rem 1.5rem; text-align: center; transition: all 0.2s; cursor: pointer; }
.upload-area.drag-over { background: #eef4ff; border-color: #2563eb; }
.upload-area input[type="file"] { display: none; }
.upload-icon { font-size: 2.5rem; margin-bottom: 0.5rem; }
.upload-area p { margin: 0.5rem 0; color: #475569; }
.upload-area .hint { font-size: 0.85rem; color: #6b7b8f; }
.btn { background: #2563eb; color: white; border: none; padding: 0.9rem 1.8rem; font-size: 1rem; font-weight: 500; border-radius: 2rem; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; gap: 0.5rem; width: 100%; box-shadow: 0 4px 6px -2px rgba(37,99,235,0.3); }
.btn:hover:not(:disabled) { background: #1d4ed8; }
.btn:disabled { background: #94a3b8; cursor: not-allowed; opacity: 0.7; }
.preview-image { max-width: 100%; border-radius: 1rem; box-shadow: 0 4px 12px rgba(0,0,0,0.08); display: block; margin: 0 auto; max-height: 220px; object-fit: contain; background: #f1f5f9; }
.canvas-wrapper { background: #0f1a2e; border-radius: 1.2rem; padding: 0.75rem; display: flex; align-items: center; justify-content: center; min-height: 260px; overflow: auto; }
canvas { display: block; max-width: 100%; height: auto; border-radius: 0.6rem; background: white; }
.meta-panel { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; background: #f1f5f9; padding: 0.8rem 1.2rem; border-radius: 2rem; font-size: 0.9rem; font-weight: 500; }
.badge { background: white; border-radius: 30px; padding: 0.25rem 1rem; }
.full-text-box { background: #0f1a2e; color: #e2e8f0; border-radius: 1.2rem; padding: 1rem 1.2rem; font-family: monospace; font-size: 0.9rem; line-height: 1.6; white-space: pre-wrap; word-break: break-word; overflow-y: auto; max-height: 150px; }
.error-message { background: #fee2e2; border-left: 6px solid #dc2626; color: #991b1b; padding: 0.9rem 1.2rem; border-radius: 0.8rem; font-size: 0.95rem; }
.spinner { display: inline-block; width: 1.1rem; height: 1.1rem; border: 2px solid rgba(255,255,255,0.3); border-radius: 50%; border-top-color: white; animation: spin 0.8s linear infinite; margin-right: 6px; }
@keyframes spin { to { transform: rotate(360deg); } }
.label { font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.5px; color: #64748b; font-weight: 600; margin-bottom: 0.3rem; }
@media (max-width: 700px) { .main { padding: 1.2rem; } .header { padding: 1rem 1.5rem; } .header h1 { font-size: 1.4rem; } }
</style>
</head>
<body>
<div class="card">
  <div class="header">
    <h1>📄 OCR 文字识别 <span>POST /ocr</span></h1>
  </div>
  <div class="main">
    <div class="left">
      <div class="upload-area" id="dropZone">
        <div class="upload-icon">📤</div>
        <p>点击或拖拽图片到此处</p>
        <p class="hint">支持 PNG, JPG, WEBP</p>
        <input type="file" id="fileInput" accept="image/*">
      </div>
      <div>
        <div class="label">原始图片预览</div>
        <img id="previewImg" class="preview-image" alt="预览图" style="display: none;" />
        <div id="previewPlaceholder" style="color:#94a3b8; font-size:0.9rem; text-align:center; padding:1rem;">尚无图片</div>
      </div>
      <button class="btn" id="recognizeBtn" disabled><span>🔍 开始识别</span></button>
      <div id="errorContainer" style="display: none;" class="error-message"></div>
    </div>
    <div class="right">
      <div class="meta-panel">
        <div>⏱️ 耗时 <span class="badge" id="elapsedMs">—</span> ms</div>
        <div>📦 识别框数量 <span class="badge" id="boxCount">—</span></div>
      </div>
      <div class="label">识别可视化（带文字框）</div>
      <div class="canvas-wrapper" id="canvasWrapper">
        <canvas id="resultCanvas"></canvas>
        <div id="canvasPlaceholder" style="color:#94a3b8; font-size:0.95rem; text-align:center; padding:2rem;">等待识别</div>
      </div>
      <div class="label">完整文本 (full_text)</div>
      <div class="full-text-box" id="fullTextBox">—</div>
    </div>
  </div>
</div>
<script>
(function() {
  var dropZone = document.getElementById('dropZone');
  var fileInput = document.getElementById('fileInput');
  var previewImg = document.getElementById('previewImg');
  var previewPlaceholder = document.getElementById('previewPlaceholder');
  var recognizeBtn = document.getElementById('recognizeBtn');
  var errorContainer = document.getElementById('errorContainer');
  var resultCanvas = document.getElementById('resultCanvas');
  var canvasPlaceholder = document.getElementById('canvasPlaceholder');
  var elapsedMsSpan = document.getElementById('elapsedMs');
  var boxCountSpan = document.getElementById('boxCount');
  var fullTextBox = document.getElementById('fullTextBox');

  var currentFile = null;
  var currentImage = null;
  var isProcessing = false;

  function showError(msg) {
    errorContainer.textContent = msg;
    errorContainer.style.display = 'block';
  }
  function clearError() {
    errorContainer.style.display = 'none';
    errorContainer.textContent = '';
  }
  function resetResults() {
    elapsedMsSpan.textContent = '—';
    boxCountSpan.textContent = '—';
    fullTextBox.textContent = '—';
    var ctx = resultCanvas.getContext('2d');
    ctx.clearRect(0, 0, resultCanvas.width, resultCanvas.height);
    resultCanvas.width = 0;
    resultCanvas.height = 0;
    resultCanvas.style.display = 'none';
    canvasPlaceholder.style.display = 'block';
  }
  function resetAll() {
    resetResults();
    previewImg.style.display = 'none';
    previewImg.src = '';
    previewPlaceholder.style.display = 'block';
    currentFile = null;
    currentImage = null;
    recognizeBtn.disabled = true;
    clearError();
    fileInput.value = '';
  }

  function loadImageFile(file) {
    if (!file) return;
    if (!file.type.startsWith('image/')) {
      showError('请选择图片文件 (png, jpg, webp 等)');
      return;
    }
    clearError();
    currentFile = file;
    var reader = new FileReader();
    reader.onload = function(e) {
      var img = new Image();
      img.onload = function() {
        if (previewImg.src && previewImg.src.indexOf('blob:') === 0) {
          URL.revokeObjectURL(previewImg.src);
        }
        var objUrl = URL.createObjectURL(file);
        previewImg.src = objUrl;
        previewImg.style.display = 'block';
        previewPlaceholder.style.display = 'none';
        currentImage = img;
        recognizeBtn.disabled = false;
        resetResults();
      };
      img.onerror = function() {
        showError('图片加载失败，请重试');
        recognizeBtn.disabled = true;
      };
      img.src = e.target.result;
    };
    reader.readAsDataURL(file);
  }

  dropZone.addEventListener('click', function() { fileInput.click(); });
  dropZone.addEventListener('dragover', function(e) {
    e.preventDefault();
    dropZone.classList.add('drag-over');
  });
  dropZone.addEventListener('dragleave', function() {
    dropZone.classList.remove('drag-over');
  });
  dropZone.addEventListener('drop', function(e) {
    e.preventDefault();
    dropZone.classList.remove('drag-over');
    var files = e.dataTransfer.files;
    if (files.length > 0) {
      loadImageFile(files[0]);
      fileInput.files = files;
    }
  });
  fileInput.addEventListener('change', function(e) {
    if (e.target.files.length > 0) {
      loadImageFile(e.target.files[0]);
    }
  });

  function drawResults(image, ocrData) {
    if (!image) { showError('图片未加载'); return; }
    if (!ocrData || !ocrData.results || !Array.isArray(ocrData.results)) {
      showError('接口返回数据格式异常');
      return;
    }
    var results = ocrData.results;
    var fullText = ocrData.full_text !== undefined && ocrData.full_text !== null ? ocrData.full_text : '';
    var elapsedMs = ocrData.elapsed_ms !== undefined ? ocrData.elapsed_ms : '—';

    elapsedMsSpan.textContent = (typeof elapsedMs === 'number') ? elapsedMs.toFixed(2) : elapsedMs;
    boxCountSpan.textContent = results.length;
    fullTextBox.textContent = fullText.trim() || '（空）';

    var canvas = resultCanvas;
    canvas.width = image.naturalWidth;
    canvas.height = image.naturalHeight;
    canvas.style.display = 'block';
    canvasPlaceholder.style.display = 'none';

    var ctx = canvas.getContext('2d');
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.drawImage(image, 0, 0, canvas.width, canvas.height);

    if (results.length > 0) {
      results.forEach(function(item) {
        var box = item.box;
        var text = item.text || '';
        var textColor = item.text_color || [0, 0, 0];
        if (!box || !Array.isArray(box) || box.length < 3) return;

        ctx.beginPath();
        ctx.moveTo(box[0][0], box[0][1]);
        for (var i = 1; i < box.length; i++) {
          ctx.lineTo(box[i][0], box[i][1]);
        }
        ctx.closePath();
        ctx.strokeStyle = '#00e0ff';
        ctx.lineWidth = Math.max(2, Math.floor(canvas.width / 400));
        ctx.shadowColor = 'rgba(0,0,0,0.7)';
        ctx.shadowBlur = 6;
        ctx.stroke();
        ctx.fillStyle = 'rgba(0,224,255,0.08)';
        ctx.fill();
        ctx.shadowBlur = 0;

        if (text.trim() !== '') {
          var minX = Infinity, minY = Infinity;
          box.forEach(function(p) {
            if (p[0] < minX) minX = p[0];
            if (p[1] < minY) minY = p[1];
          });
          var fontSize = Math.max(12, Math.floor(canvas.width / 50));
          ctx.font = 'bold ' + fontSize + 'px "Segoe UI", Roboto, sans-serif';
          ctx.textBaseline = 'top';
          var textMetrics = ctx.measureText(text);
          var textWidth = textMetrics.width;
          var padding = 6;
          var bgW = textWidth + padding * 2;
          var bgH = fontSize + padding * 1.5;
          var drawX = minX;
          var drawY = minY - fontSize - padding;
          if (drawY < 0) drawY = minY + 4;

          ctx.shadowBlur = 0;
          ctx.fillStyle = 'rgba(0,20,40,0.85)';
          ctx.beginPath();
          if (ctx.roundRect) {
            ctx.roundRect(drawX, drawY, bgW, bgH, 4);
          } else {
            ctx.rect(drawX, drawY, bgW, bgH);
          }
          ctx.fill();

          var colorStr = '#ffffff';
          if (Array.isArray(textColor) && textColor.length === 3) {
            var r = Math.min(255, Math.max(0, textColor[0]));
            var g = Math.min(255, Math.max(0, textColor[1]));
            var b = Math.min(255, Math.max(0, textColor[2]));
            var brightness = (r * 0.299 + g * 0.587 + b * 0.114);
            colorStr = brightness < 100 ? '#ffffff' : 'rgb(' + r + ',' + g + ',' + b + ')';
          }
          ctx.fillStyle = colorStr;
          ctx.shadowColor = 'rgba(0,0,0,0.8)';
          ctx.shadowBlur = 4;
          ctx.fillText(text, drawX + padding, drawY + padding / 1.5);
          ctx.shadowBlur = 0;
        }
      });
    } else {
      ctx.font = 'bold 24px sans-serif';
      ctx.fillStyle = 'rgba(200,0,0,0.8)';
      ctx.shadowBlur = 8;
      ctx.shadowColor = 'black';
      ctx.fillText('未检测到文字', 30, 80);
      ctx.shadowBlur = 0;
    }
  }

  async function callOcrApi(file) {
    var formData = new FormData();
    formData.append('image', file);
    var response = await fetch('/ocr', {
      method: 'POST',
      body: formData
    });
    if (!response.ok) {
      var errorText = '';
      try {
        var errorJson = await response.json();
        errorText = errorJson.detail || JSON.stringify(errorJson);
      } catch (e) {
        errorText = await response.text();
      }
      throw new Error('OCR 请求失败 (' + response.status + '): ' + (errorText || '无详情'));
    }
    return await response.json();
  }

  recognizeBtn.addEventListener('click', async function() {
    if (isProcessing) return;
    if (!currentFile) { showError('请先选择图片'); return; }
    if (!currentImage) { showError('图片尚未加载完成，请稍候'); return; }
    isProcessing = true;
    recognizeBtn.disabled = true;
    var originalBtnHtml = recognizeBtn.innerHTML;
    recognizeBtn.innerHTML = '<span style="display:flex;align-items:center;justify-content:center;"><span class="spinner"></span> 识别中…</span>';
    clearError();
    canvasPlaceholder.style.display = 'block';
    resultCanvas.style.display = 'none';
    elapsedMsSpan.textContent = '—';
    boxCountSpan.textContent = '—';
    fullTextBox.textContent = '—';

    try {
      var data = await callOcrApi(currentFile);
      console.log('OCR 响应:', data);
      if (!data || !Array.isArray(data.results)) {
        throw new Error('接口返回格式不正确: 缺少 results 数组');
      }
      drawResults(currentImage, data);
    } catch (err) {
      console.error('OCR 失败:', err);
      showError(err.message || '识别请求失败，请检查网络或接口');
      resultCanvas.style.display = 'none';
      canvasPlaceholder.style.display = 'block';
    } finally {
      recognizeBtn.innerHTML = originalBtnHtml;
      recognizeBtn.disabled = false;
      isProcessing = false;
    }
  });

  resetAll();
  window.addEventListener('beforeunload', function() {
    if (previewImg.src && previewImg.src.indexOf('blob:') === 0) {
      URL.revokeObjectURL(previewImg.src);
    }
  });
})();
</script>
</body>
</html>`
