// cmd/cli/main.go
package main

import (
  "fmt"
  "log"
  "os"
  "text/template"
  "os/exec"
  "bufio"
  "strings"
)
func Template() string {
  return `
<?xml version='1.0' encoding='UTF-8'?>
 <!DOCTYPE plist PUBLIC \"-//Apple Computer//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\" >
 <plist version='1.0'>
   <dict>
     <key>Label</key><string>{{.Label}}</string>
     <key>WorkingDirectory</key><string>{{.ConfigDir}}</string>
     <key>StandardOutPath</key><string>/tmp/{{.Label}}.out.log</string>
     <key>StandardErrorPath</key><string>/tmp/{{.Label}}.err.log</string>
     <key>KeepAlive</key><{{.KeepAlive}}/>
     <key>RunAtLoad</key><{{.RunAtLoad}}/>
     <key>ProgramArguments</key>
     <array>
        <string>{{.Program}}</string>
        <string>-remote-ip</string>
        <string>{{.RemoteIP}}</string>
        <string>-config-dir</string>
        <string>{{.ConfigDir}}</string>
        {{if .RemoteUser}}<string>-remote-user</string> 
        <string>{{.RemoteUser}}</string>{{end}}
        {{if .RemotePath}}<string>-remote-path {{.RemotePath}}</string>{{end}}
     </array>
     <key>EnvironmentVariables</key>
      <dict>
        <key>PATH</key>
        <string>{{.PATH}}</string>
      </dict>
   </dict>
</plist>
`
}
type tdata struct {
  Label     string
  Program   string
  KeepAlive bool
  RunAtLoad bool
  RemoteIP  string
  RemoteUser  string
  RemotePath  string
  ConfigDir string
  PATH string
}
func create(plistPath string, data tdata) {
  f, err := os.OpenFile(plistPath, os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
    log.Fatalf("os.Open failed: %s", err)
  }
  err = f.Truncate(0)
  t := template.Must(template.New("launchdConfig").Parse(Template()))
  err = t.Execute(f, data)
  if err != nil {
          log.Fatalf("Template generation failed: %s", err)
  }
  f.Close()
}
func load(plistPath string) {
  unload(plistPath)
  cmd := exec.Command("launchctl", "load", plistPath)
  // show rsync's output
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr

  cmd.Run()
}
func unload(plistPath string) {
  cmd := exec.Command("launchctl", "unload", plistPath)
  // show rsync's output
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr

  cmd.Run()
}
func main() {
  reader := bufio.NewReader(os.Stdin)
  fmt.Print("Remote IP: ")
  remoteIP, _ := reader.ReadString('\n')
  remoteIP = strings.TrimSpace(remoteIP)

  dir, e := os.Getwd()
  if e != nil {
    log.Fatal(e)
  }

  data := tdata{
    Label:     "me.anhtuan.remote-work-env",
    Program:   fmt.Sprintf("%s/remote-work-env", dir),
    KeepAlive: true,
    RunAtLoad: true,
    RemoteIP: remoteIP,
    ConfigDir: dir,
    PATH: os.Getenv("PATH"),
  }

  plistPath := fmt.Sprintf("%s/Library/LaunchAgents/%s.plist", os.Getenv("HOME"), data.Label)
  create(plistPath, data)
  load(plistPath)
}
