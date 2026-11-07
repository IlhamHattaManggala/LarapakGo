import os
import base64
from PIL import Image

src_path = r"d:\Ilham Hatta Manggala\Project Gratis\Ngapakin-Bos\ngapakin\public\favicon.png"
dst_dir = r"d:\Ilham Hatta Manggala\Project Gratis\Ngapakin-Bos\LarapakGo\editors\vscode\icons"

im = Image.open(src_path)
bbox = im.getbbox()
cropped = im.crop(bbox)

# Create a square canvas centered
max_dim = max(cropped.width, cropped.height)
square = Image.new("RGBA", (max_dim, max_dim), (0, 0, 0, 0))
offset = ((max_dim - cropped.width) // 2, (max_dim - cropped.height) // 2)
square.paste(cropped, offset)

# Save high-res PNGs
square.save(os.path.join(dst_dir, "ngpk.png"))
square.save(os.path.join(dst_dir, "nview.png"))
square.save(os.path.join(dst_dir, "favicon.png"))

# Also read original base64 and create crisp SVG wrapping it
with open(os.path.join(dst_dir, "ngpk.png"), "rb") as f:
    b64 = base64.b64encode(f.read()).decode("utf-8")

svg_content = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {max_dim} {max_dim}" width="32" height="32">
  <image href="data:image/png;base64,{b64}" width="{max_dim}" height="{max_dim}"/>
</svg>"""

with open(os.path.join(dst_dir, "ngpk.svg"), "w", encoding="utf-8") as f:
    f.write(svg_content)

with open(os.path.join(dst_dir, "nview.svg"), "w", encoding="utf-8") as f:
    f.write(svg_content)

print("[OK] Successfully generated icons from favicon.png")
