; -*- coding: utf-8 -*-
Unicode true
!include "MUI2.nsh"
!include "LogicLib.nsh"
!include "x64.nsh"
!include "WinVer.nsh"
!include "${INVENTORY}"

Name "shadowinput"
OutFile "${OUTPUT}"
InstallDir "$LOCALAPPDATA\Programs\shadowinput"
RequestExecutionLevel user
SetCompressor /SOLID lzma
SetCompressorDictSize 32
CRCCheck on
ShowInstDetails show
ShowUninstDetails show
VIProductVersion "${NUMERIC_VERSION}"
VIAddVersionKey "ProductName" "shadowinput"
VIAddVersionKey "FileDescription" "shadowinput installer"
VIAddVersionKey "FileVersion" "${VERSION}"
VIAddVersionKey "LegalCopyright" "shadowinput contributors; based on AstraFoundry/Arion"

!define UNINSTALL_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\shadowinput"
!define MUI_ABORTWARNING
!define MUI_ICON "${PAYLOAD}/shadowinput.ico"
!define MUI_UNICON "${PAYLOAD}/shadowinput.ico"
!define MUI_WELCOMEPAGE_TITLE "安装 shadowinput"
!define MUI_WELCOMEPAGE_TEXT "可切换的 AI 拼音输入法。$\r$\n$\r$\n将安装到当前用户的应用目录，并添加开始菜单入口，主程序无需管理员权限；在应用内启用系统输入法时，需要单独的管理员授权。$\r$\n$\r$\n安装包包含本地推理运行文件；模型由你在应用内另行下载。升级前请从托盘退出 shadowinput。$\r$\n$\r$\n${VERSION} · 未签名开发预览版"
!define MUI_FINISHPAGE_RUN "$INSTDIR\shadowinput.exe"
!define MUI_FINISHPAGE_RUN_TEXT "打开 shadowinput"
!define MUI_FINISHPAGE_RUN_NOTCHECKED
!define MUI_UNCONFIRMPAGE_TEXT_TOP "将移除 shadowinput 程序和开始菜单入口。$\r$\n$\r$\n你下载的模型、个人风格和账户设置会保留。卸载前请从托盘退出应用。"
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "SimpChinese"

!macro Initialize PREFIX
Function ${PREFIX}.onInit
  SetShellVarContext current
  SetRegView 64
  ${IfNot} ${IsNativeAMD64}
    MessageBox MB_OK|MB_ICONSTOP "此版本需要 Windows x64（Intel / AMD 64 位）。" /SD IDOK
    SetErrorLevel 2
    Abort
  ${EndIf}
  ${IfNot} ${AtLeastWin10}
    MessageBox MB_OK|MB_ICONSTOP "此版本需要 Windows 10 或更新版本。" /SD IDOK
    SetErrorLevel 2
    Abort
  ${EndIf}
  ; Keep the handle open until process exit to serialize setup/uninstall.
  System::Call 'kernel32::CreateMutexW(p 0, i 0, w "Local\shadowinput-setup") p.r0 ?e'
  Pop $1
  ${If} $0 == 0
  ${OrIf} $1 == 183
    MessageBox MB_OK|MB_ICONSTOP "另一个安装或卸载程序正在运行，请关闭后再试。" /SD IDOK
    SetErrorLevel 2
    Abort
  ${EndIf}
FunctionEnd
!macroend
!insertmacro Initialize ""
!insertmacro Initialize "un"

; Check every native payload before writing/removing any files. Never kill a
; running app or its model worker; the user can quit from the tray and retry.
!macro FileLockCheck PREFIX
Function ${PREFIX}CheckFileUnlocked
  Exch $0
  Push $1
  IfFileExists "$0" 0 done
  System::Call 'kernel32::CreateFileW(w r0, i 0x40000000, i 0, p 0, i 3, i 0, p 0) p.r1'
  ${If} $1 == -1
    MessageBox MB_OK|MB_ICONSTOP "无法更新程序文件。请从托盘退出 shadowinput，确认安装目录可写后重试。$\r$\n$0" /SD IDOK
    SetErrorLevel 2
    Abort
  ${EndIf}
  System::Call 'kernel32::CloseHandle(p r1)'
done:
  Pop $1
  Pop $0
FunctionEnd
!macroend
!insertmacro FileLockCheck ""
!insertmacro FileLockCheck "un."

Section "shadowinput"
  !insertmacro CheckPayload ""
  SetOutPath "$INSTDIR"
  SetOverwrite on
  File /r "${PAYLOAD}/*"
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  CreateShortcut "$SMPROGRAMS\shadowinput.lnk" "$INSTDIR\shadowinput.exe"
  WriteRegStr HKCU "Software\Classes\shadowinput" "" "URL:shadowinput"
  WriteRegStr HKCU "Software\Classes\shadowinput" "URL Protocol" ""
  WriteRegStr HKCU "Software\Classes\shadowinput\shell\open\command" "" '$\"$INSTDIR\shadowinput.exe$\" $\"%1$\"'
  WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayName" "shadowinput"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "Publisher" "shadowinput contributors"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayIcon" "$INSTDIR\shadowinput.exe"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "UninstallString" '$\"$INSTDIR\Uninstall.exe$\"'
  WriteRegStr HKCU "${UNINSTALL_KEY}" "QuietUninstallString" '$\"$INSTDIR\Uninstall.exe$\" /S'
  WriteRegStr HKCU "${UNINSTALL_KEY}" "URLInfoAbout" "https://shadowinput.com/"
  WriteRegDWORD HKCU "${UNINSTALL_KEY}" "EstimatedSize" ${INSTALLED_KIB}
  WriteRegDWORD HKCU "${UNINSTALL_KEY}" "NoModify" 1
  WriteRegDWORD HKCU "${UNINSTALL_KEY}" "NoRepair" 1
SectionEnd

Section "Uninstall"
  !insertmacro CheckPayload "un."
  ; Remove only our registered TSF identity before removing its maintenance helper.
  ExecWait '"$INSTDIR\shadowinput.exe" --remove-input-method' $0
  ${If} $0 != 0
    MessageBox MB_OK|MB_ICONSTOP "系统输入法尚未移除。请切换到其他输入法、关闭使用它的应用，再次卸载并完成管理员授权。"
    SetErrorLevel 1
    Abort
  ${EndIf}
  !insertmacro RemovePayload
  Delete "$SMPROGRAMS\shadowinput.lnk"
  DeleteRegKey HKCU "${UNINSTALL_KEY}"
  DeleteRegKey HKCU "Software\Classes\shadowinput"
  Delete "$INSTDIR\Uninstall.exe"
  ; Do not recurse: unknown files, profiles and model downloads are preserved.
  RMDir "$INSTDIR"
SectionEnd
