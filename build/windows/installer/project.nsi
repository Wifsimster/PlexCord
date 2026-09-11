Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
##
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "MyProject" # Default "{{.Name}}"
## !define INFO_COMPANYNAME    "MyCompany" # Default "{{.Info.CompanyName}}"
## !define INFO_PRODUCTNAME    "MyProduct" # Default "{{.Info.ProductName}}"
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "{{.Info.ProductVersion}}"
## !define INFO_COPYRIGHT      "Copyright" # Default "{{.Info.Copyright}}"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
####

####
## PlexCord installs per user, not machine-wide, and that is load-bearing rather
## than a preference: PlexCord updates itself by replacing its own executable in
## place (internal/version/update.go). Under Program Files that write needs
## elevation, so every background update would have to raise a UAC prompt — or
## stop being automatic. Installed under the user's own profile it stays a plain
## file write, and the update lands silently like it does for the portable build.
##
## Consequences of "user" that the rest of this file has to honor:
##   - wails.setShellContext resolves shortcuts to the current user.
##   - wails_tools.nsh's wails.writeUninstaller writes to HKLM, which a
##     non-elevated installer cannot do, so the uninstall entry below is written
##     to HKCU by hand instead. Do not switch back to that macro without also
##     switching back to an admin install.
####
!define REQUEST_EXECUTION_LEVEL "user"

# Without this the key would default to "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}",
# i.e. "PlexCordPlexCord".
!define UNINST_KEY_NAME "PlexCord"

####
## Include the wails tools
####
!include "wails_tools.nsh"

####
## Uninstall registration, per user. Mirrors wails.writeUninstaller/
## wails.deleteUninstaller from wails_tools.nsh, with HKCU in place of HKLM.
####
!macro plexcord.writeUninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"

    SetRegView 64

    WriteRegStr HKCU "${UNINST_KEY}" "Publisher" "${INFO_COMPANYNAME}"
    WriteRegStr HKCU "${UNINST_KEY}" "DisplayName" "${INFO_PRODUCTNAME}"
    WriteRegStr HKCU "${UNINST_KEY}" "DisplayVersion" "${INFO_PRODUCTVERSION}"
    WriteRegStr HKCU "${UNINST_KEY}" "DisplayIcon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr HKCU "${UNINST_KEY}" "InstallLocation" "$INSTDIR"
    WriteRegStr HKCU "${UNINST_KEY}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKCU "${UNINST_KEY}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
    WriteRegDWORD HKCU "${UNINST_KEY}" "NoModify" 1
    WriteRegDWORD HKCU "${UNINST_KEY}" "NoRepair" 1

    ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
    IntFmt $0 "0x%08X" $0
    WriteRegDWORD HKCU "${UNINST_KEY}" "EstimatedSize" "$0"

    # Where the next installer run should default to, so an upgrade lands on
    # top of the existing install instead of beside it.
    WriteRegStr HKCU "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallDir" "$INSTDIR"
!macroend

!macro plexcord.deleteUninstaller
    Delete "$INSTDIR\uninstall.exe"

    SetRegView 64

    DeleteRegKey HKCU "${UNINST_KEY}"
    DeleteRegKey HKCU "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!macroend

####
## Stop a running PlexCord before touching its files: the executable is locked
## while it runs, and PlexCord is built to sit in the tray, so it is usually
## running when an upgrade is installed. taskkill without /F asks a GUI app to
## close, which PlexCord's "Minimize to tray" setting turns into "hide the
## window and keep running", hence the forced pass after a short grace period.
####
!macro plexcord.stopRunningInstance
    DetailPrint "Closing PlexCord if it is running..."
    nsExec::Exec 'taskkill /IM "${PRODUCT_EXECUTABLE}"'
    Pop $0
    Sleep 2000
    nsExec::Exec 'taskkill /F /IM "${PRODUCT_EXECUTABLE}"'
    Pop $0
    Sleep 500
!macroend

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

!insertmacro MUI_LANGUAGE "English" # Set the Language of the installer

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
# Per-user install location, the one PlexCord can write to without elevation.
# $LOCALAPPDATA\Programs is where Windows expects a per-user app to live.
InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
# An upgrade reuses the directory the previous install recorded.
InstallDirRegKey HKCU "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallDir"
ShowInstDetails show # This will always show the installation details.

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro plexcord.stopRunningInstance

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro plexcord.writeUninstaller
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    !insertmacro plexcord.stopRunningInstance

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    SetRegView 64
    # PlexCord registers itself here when "Start on login" is on
    # (internal/platform/autostart_windows.go); leaving the value behind would
    # point Windows at a deleted executable on every boot.
    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "PlexCord"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro plexcord.deleteUninstaller
SectionEnd
