import os
import shutil
import tempfile
from fastapi import FastAPI, UploadFile, File, HTTPException
from paddleocr import PaddleOCR

app = FastAPI()

# Initialize global OCR model
# Load once at startup as requested
print("Loading PaddleOCR model...")
# Switch to 'en' as it often handles receipt fonts/digits better than 'id'
ocr_model = PaddleOCR(use_angle_cls=True, lang='en') 
print("PaddleOCR model loaded successfully.")

@app.get("/")
def read_root():
    return {"message": "OCR Service is running"}

@app.post("/scan")
async def scan_image(file: UploadFile = File(...)):
    # Create a temporary file to save the uploaded image
    # We preserve the extension just in case, though PaddleOCR handles most formats
    extension = os.path.splitext(file.filename)[1]
    if not extension:
        extension = ".jpg" # Default fallback
        
    with tempfile.NamedTemporaryFile(delete=False, suffix=extension) as tmp:
        shutil.copyfileobj(file.file, tmp)
        tmp_path = tmp.name

    try:
        # DEBUG: Log file details
        file_size = os.path.getsize(tmp_path)
        print(f"DEBUG: Received file: {file.filename}, Size: {file_size} bytes, Path: {tmp_path}", flush=True)
        
        # DEBUG: Check first bytes to see if it is an image or HTML (error page)
        with open(tmp_path, 'rb') as f_head:
            head = f_head.read(20)
            print(f"DEBUG: File header: {head}", flush=True)

        # Run OCR
        result = ocr_model.ocr(tmp_path)
        print(f"DEBUG: OCR Raw Result: {result}", flush=True)

        full_text = ""
        # PaddleOCR returns a list of results (one per image). 
        # Since we passed one path, we look at result[0].
        # result structure: [ [ [box], (text, score) ], ... ]
        if result and result[0]:
            # Check if result[0] is a dict (new structure seen in logs) or list (old structure)
            if isinstance(result[0], dict):
                # Structure: {'rec_texts': [...], 'rec_scores': [...], ...}
                data = result[0]
                rec_texts = data.get('rec_texts', [])
                rec_scores = data.get('rec_scores', [])
                
                texts = []
                for i, text in enumerate(rec_texts):
                    score = rec_scores[i] if i < len(rec_scores) else 0.0
                    if score < 0.8:
                        print(f"DEBUG: Low confidence ({score:.2f}): {text}", flush=True)
                    texts.append(text)
                full_text = "\n".join(texts)
                
            elif isinstance(result[0], list):
                # Old Structure: [ [ [box], (text, score) ], ... ]
                texts = []
                for line in result[0]:
                    text = line[1][0]
                    score = line[1][1]
                    if score < 0.8:
                        print(f"DEBUG: Low confidence ({score:.2f}): {text}", flush=True)
                    texts.append(text)
                full_text = "\n".join(texts)

            print(f"DEBUG: Full extracted text: {full_text}", flush=True)

        return {"text": full_text}

    except Exception as e:
        # Log error in production scenario
        print(f"ERROR executing OCR: {e}", flush=True) # Added explicit print
        raise HTTPException(status_code=500, detail=str(e))
        
    finally:
        # Ensure temporary file is always deleted
        if os.path.exists(tmp_path):
            os.remove(tmp_path)
