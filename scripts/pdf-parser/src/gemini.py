import logging
from google import genai
from google.genai.types import Part, UserContent, GenerateContentConfig
from PIL.Image import Image as ImageClass
from io import BytesIO
from PIL.Image import Image as ImageClass
from config import GCP_PROJECT, GCP_REGION, GEMINI_LITE, SYSTEM_INSTRUCTION

logger = logging.getLogger(__name__)

def pil_image_to_bytes(image: ImageClass) -> bytes:
        buffer = BytesIO()
        image.save(buffer, format="PNG")
        return buffer.getvalue()

def get_gemini_response(ocr_output: str, image: ImageClass, page_num: int) -> str:
    client = genai.Client(
        vertexai=True,
        project=GCP_PROJECT,
        location=GCP_REGION
    )

    contents = UserContent(
        parts = [
            Part(text="Here's the OCR output for this page:"),
            Part(text=str(ocr_output)),
            Part(text="Here's the image:"),
            Part.from_bytes(
                data=pil_image_to_bytes(image),
                mime_type="image/png",
            ),
        ]
    )
    config = GenerateContentConfig(
        system_instruction = SYSTEM_INSTRUCTION
    )

    logger.info(msg=f"Sending Page #{page_num} to Gemini...")
    response = client.models.generate_content(
        model=str(GEMINI_LITE),
        contents=contents,
        config=config,
    )

    logger.info(msg=f"Successfully received response")
    return str(response.text)
