; =========================================================================
; LarapakGo & NgapakIn Language Engine - Inno Setup Script
; Versi: 1.0.0
; Lisensi: MIT
; =========================================================================

#define MyAppName "LarapakGo"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "Ilham Hatta Manggala"
#define MyAppURL "https://larapak.my.id"
#define MyAppExeName "larapak.exe"

[Setup]
AppId={{8B264A7E-D93F-4182-9B7C-4C82E874D11E}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL=https://github.com/IlhamHattaManggala/LarapakGo
AppUpdatesURL=https://github.com/IlhamHattaManggala/LarapakGo/releases
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
LicenseFile=..\LICENSE
OutputDir=..\dist
OutputBaseFilename=LarapakGo-Setup-v{#MyAppVersion}
SetupIconFile=app.ico
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
ChangesEnvironment=yes
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
UninstallDisplayIcon={app}\{#MyAppExeName}

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "..\dist\larapak-v1.0.0-windows-amd64.exe"; DestDir: "{app}"; DestName: "larapak.exe"; Flags: ignoreversion
Source: "..\dist\larapak-v1.0.0-windows-amd64.exe"; DestDir: "{app}"; DestName: "ngapakin.exe"; Flags: ignoreversion
Source: "..\LICENSE"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\README.md"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Larapak REPL (Terminal Interaktif)"; Filename: "{app}\larapak.exe"; Parameters: "repl"; IconFilename: "{app}\larapak.exe"
Name: "{group}\Dokumentasi Resmi Larapak"; Filename: "{#MyAppURL}"
Name: "{group}\Uninstall {#MyAppName}"; Filename: "{uninstallexe}"

[Code]
const
    EnvironmentKey = 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';

procedure AddToPath();
var
    Paths: String;
    AppPath: String;
begin
    AppPath := ExpandConstant('{app}');
    if IsAdminInstallMode then
    begin
        if RegQueryStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths) then
        begin
            if Pos(';' + AppPath + ';', ';' + Paths + ';') = 0 then
            begin
                Paths := Paths + ';' + AppPath;
                RegWriteStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths);
            end;
        end;
    end
    else
    begin
        if RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', Paths) then
        begin
            if Pos(';' + AppPath + ';', ';' + Paths + ';') = 0 then
            begin
                Paths := Paths + ';' + AppPath;
                RegWriteStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', Paths);
            end;
        end
        else
        begin
            RegWriteStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', AppPath);
        end;
    end;
end;

procedure RemoveFromPath();
var
    Paths: String;
    AppPath: String;
    P: Integer;
begin
    AppPath := ExpandConstant('{app}');
    if IsAdminInstallMode then
    begin
        if RegQueryStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths) then
        begin
            P := Pos(';' + AppPath, Paths);
            if P > 0 then
            begin
                Delete(Paths, P, Length(';' + AppPath));
                RegWriteStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths);
            end
            else
            begin
                P := Pos(AppPath + ';', Paths);
                if P > 0 then
                begin
                    Delete(Paths, P, Length(AppPath + ';'));
                    RegWriteStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths);
                end
                else if Paths = AppPath then
                begin
                    RegWriteStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', '');
                end;
            end;
        end;
    end
    else
    begin
        if RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', Paths) then
        begin
            P := Pos(';' + AppPath, Paths);
            if P > 0 then
            begin
                Delete(Paths, P, Length(';' + AppPath));
                RegWriteStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', Paths);
            end
            else
            begin
                P := Pos(AppPath + ';', Paths);
                if P > 0 then
                begin
                    Delete(Paths, P, Length(AppPath + ';'));
                    RegWriteStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', Paths);
                end
                else if Paths = AppPath then
                begin
                    RegWriteStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', '');
                end;
            end;
        end;
    end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
    if CurStep = ssPostInstall then
    begin
        AddToPath();
    end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
    if CurUninstallStep = usPostUninstall then
    begin
        RemoveFromPath();
    end;
end;

