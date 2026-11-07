import os
import zipfile

content_types = """<?xml version="1.0" encoding="utf-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="json" ContentType="application/json"/>
  <Default Extension="md" ContentType="text/markdown"/>
  <Default Extension="vsixmanifest" ContentType="text/xml"/>
  <Default Extension="txt" ContentType="text/plain"/>
  <Default Extension="svg" ContentType="image/svg+xml"/>
</Types>
"""

vsix_manifest = """<?xml version="1.0" encoding="utf-8"?>
<PackageManifest Version="2.0.0" xmlns="http://schemas.microsoft.com/developer/vsx-schema/2011" xmlns:d="http://schemas.microsoft.com/developer/vsx-schema-design/2011">
  <Metadata>
    <Identity Id="larapak-vscode" Version="1.2.2" Language="en-US" Publisher="IlhamHattaManggala"/>
    <DisplayName>Larapak &amp; NgapakIn Language Support</DisplayName>
    <Description xml:space="preserve">Syntax highlighting, file icons, snippets, and language support for NgapakIn (.ngpk) scripts and Sawangan (.nview) templates in LarapakGo.</Description>
    <Icon>extension/icons/ngpk.svg</Icon>
    <Categories>Programming Languages,Snippets</Categories>
  </Metadata>
  <Installation>
    <InstallationTarget Id="Microsoft.VisualStudio.Code"/>
  </Installation>
  <Dependencies/>
  <Assets>
    <Asset Type="Microsoft.VisualStudio.Code.Manifest" Path="extension/package.json" Addressable="true"/>
    <Asset Type="Microsoft.VisualStudio.Services.Content.Details" Path="extension/README.md" Addressable="true"/>
    <Asset Type="Microsoft.VisualStudio.Services.Content.License" Path="extension/LICENSE" Addressable="true"/>
    <Asset Type="Microsoft.VisualStudio.Services.Icons.Default" Path="extension/icons/ngpk.svg" Addressable="true"/>
  </Assets>
</PackageManifest>
"""

output_vsix = "larapak-vscode-1.2.2.vsix"
root_dir = os.path.dirname(os.path.abspath(__file__))

with zipfile.ZipFile(os.path.join(root_dir, output_vsix), "w", zipfile.ZIP_DEFLATED) as z:
    z.writestr("[Content_Types].xml", content_types)
    z.writestr("extension.vsixmanifest", vsix_manifest)
    
    files_to_pack = [
        "package.json",
        "README.md",
        "LICENSE",
        "language-configuration-ngpk.json",
        "language-configuration-nview.json",
        os.path.join("syntaxes", "ngapak.tmLanguage.json"),
        os.path.join("syntaxes", "nview.tmLanguage.json"),
        os.path.join("snippets", "ngapak.json"),
        os.path.join("snippets", "nview.json"),
    ]

    icons_dir = os.path.join(root_dir, "icons")
    if os.path.exists(icons_dir):
        for icon_file in os.listdir(icons_dir):
            files_to_pack.append(os.path.join("icons", icon_file))
    
    for f in files_to_pack:
        file_path = os.path.join(root_dir, f)
        arcname = "extension/" + f.replace("\\", "/")
        z.write(file_path, arcname)

print(f"[OK] Sukses nggawe {output_vsix}")
