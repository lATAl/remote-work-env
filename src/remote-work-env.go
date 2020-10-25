package main

import (
  "fmt"
  "log"
  "io/ioutil"
  "os"
  "os/exec"
  "os/user"
  "strings"
  "regexp"
  "bufio"
  "gopkg.in/yaml.v2"
  "strconv"
  "flag"
  "os/signal"
  "syscall"
  "path/filepath"
)
var debug = false
var debugoutput = false
var fw_opt_template = "-L %s:127.0.0.1:%s"

type project struct {
    name string
    path string
    ports []string
}

type DockerConfig struct {
  Version  string
  Services map[string]DockerService
}

type DockerService struct {
  Ports []string
}

func makeRangeStr(min, max int) []string {
    a := make([]string, max-min+1)
    for i := range a {
        a[i] = strconv.Itoa(min + i)
    }
    return a
}

func load_project_ports(p project) project {
  docker_compose := fmt.Sprintf("%s/docker-compose.yml", p.path)
  if fileExists(docker_compose) {
    dcc, err := ioutil.ReadFile(docker_compose)
    check(err)
    data := &DockerConfig{}
    err = yaml.Unmarshal([]byte(dcc), &data)
    if err != nil {
      log.Fatalf("error: %v", err)
    }
    for _, service := range data.Services {
      for _, port := range service.Ports {
        port = strings.Split(port, ":")[0]
        ports := strings.Split(port, "-")
        if len(ports) > 1 {
          start, err := strconv.Atoi(ports[0])
          check(err)
          end, err := strconv.Atoi(ports[1])
          check(err)
          if start < end {
            makeRangeStr(start, end)
          }
        }
        p.ports = append(p.ports, ports...)
      }
    }
  }
  return p
}

func load_project(base_path string, name string) project {
  name_w_ports := strings.Split(name, ":")
  p := project{name: name_w_ports[0]}
  if len(name_w_ports) > 1 {
    p.ports = name_w_ports[1:]
  }
  p.path = fmt.Sprintf("%s/%s", base_path, p.name)
  return load_project_ports(p)
}
func fw_port_opts(port string) string {
  return fmt.Sprintf(fw_opt_template, port, port)
}
func fw_port(remoteUser string, remoteIP string, projects []project) {
  ports := []string{}
  for _, project := range projects{
    ports = append(ports, project.ports...)
  }
  args := []string{
    "-M 0",
    "-o", "ServerAliveInterval 30",
    "-o", "ServerAliveCountMax 3",
    "-N",
    "-R 9999:127.0.0.1:9999", //clipboard
  }
  for _, port := range ports{
    args = append(args, fw_port_opts(port))
  }

  socket_file := "/tmp/ssh/dev-server-socket"
  args = append(args, fmt.Sprintf("-S %s", socket_file))

  args = append(args, fmt.Sprintf("%s@%s", remoteUser, remoteIP))

  cmd := exec.Command("autossh", args...)
  // show rsync's output
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  if debug || debugoutput {
    fmt.Printf("Debug: \n%v\n", cmd.Args)
    if debug {
      return
    }
  }

  fmt.Printf("FW PORT %v.\n", ports)
  cmd.Run()
}

func main() {
  user, err := user.Current()
  remoteUserPtr := flag.String("remote-user", user.Username, "Remote username")
  remotePathPtr := flag.String("remote-path", "~/dev/", "Remote path")
  remoteIPptr := flag.String("remote-ip", "", "Remote IP")
  configDirPtr := flag.String("config-dir", "", "Config dir")
  workingDirPtr := flag.String("working-dir", "", "Config dir")
  flag.Parse()
  if *remoteIPptr == "" {
    flag.PrintDefaults()
    os.Exit(1)
  }
  if *configDirPtr == "" {
    flag.PrintDefaults()
    os.Exit(1)
  }

  remoteUser := fmt.Sprintf("%s", *remoteUserPtr)
  remotePath := fmt.Sprintf("%s", *remotePathPtr)
  remoteIP := fmt.Sprintf("%s", *remoteIPptr)
  configDir := fmt.Sprintf("%s", *configDirPtr)

  workingDir := fmt.Sprintf("%s", *workingDirPtr)
  if workingDir == "" {
    workingDir = filepath.Dir(configDir)
  }

  sigs := make(chan os.Signal, 1)
  signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

  go func() {
    <-sigs
    os.Exit(0)
  }()

  project_name_text, err := ioutil.ReadFile(fmt.Sprintf("%s/project_name", configDir))
  check(err)
  var project_names = strings.Split(strings.TrimSpace(string(project_name_text)), "\n")
  
  projects := []project{}
  for _, project_name := range project_names{
    // project_path := fmt.Sprintf("../%s", project_name)
    projects = append(projects, load_project(workingDir, project_name))
    // rsync(project_path, remoteIP, remotePath)
    // go watch(project_path, changes)
    // rsync(project_path, "113.20.119.39", remotePath)
  }

  go fw_port(remoteUser, remoteIP, projects)

  changes := make(chan string)
  done := make(chan bool)
  go func() {
    for {
      select {
        case msg1 := <-changes:
          fmt.Println(msg1, "has change")
          rsync(msg1, remoteUser, remoteIP, remotePath)
      }
    }
  }()

  for _, project := range projects{
    go rsync(project.path, remoteUser, remoteIP, remotePath)
    go watch(project.path, changes)
  }
  
  <-done
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}
func Reject(vs []string, f func(string) bool) []string {
    vsf := make([]string, 0)
    for _, v := range vs {
        if !f(v) {
            vsf = append(vsf, v)
        }
    }
    return vsf
}
func Map(vs []string, f func(string) string) []string {
    vsm := make([]string, len(vs))
    for i, v := range vs {
        vsm[i] = f(v)
    }
    return vsm
}

func build_exclude(project_path string) []string {
  gitignore := fmt.Sprintf("%s/.gitignore", project_path)
  if (!fileExists(gitignore)) {return []string{}}
  ignoreContent, err := ioutil.ReadFile(gitignore)
  check(err)
  var ignores = strings.Split(strings.TrimSpace(string(ignoreContent)), "\n")
  ignores = Reject(ignores, func(v string) bool {
    matchComment, _ := regexp.MatchString("^#.*", v)
    if (matchComment) {return true}
    matchBs, _ := regexp.MatchString("^\\#", v)
    if (matchBs) {return true}
    matchEmpty, _ := regexp.MatchString("^$", v)
    if (matchEmpty) {return true}
    matchGoooleKey, _ := regexp.MatchString("^/priv/google/", v)
    if (matchGoooleKey) {return true}
    if (v == "\\#*\\#") {return true}
    if (v == ".\\#*") {return true}
    return false
  })
  ignores = Map(ignores, func(v string) string {
    if (strings.HasPrefix(v, "/")) {
      v = v[1:]
      if v[len(v)-1:] != "/" {
        v = v + "/"
      }
    }
    // if strings.Contains(v, "*") {
    //   v = "'" + v + "'"
    // }
    return fmt.Sprintf("--exclude=%s", v)
  })
  return ignores
}
func rsync(project_path string, remoteUser string, remoteIP string, remotePath string) error {
  remote := fmt.Sprintf("%s@%s:%s", remoteUser, remoteIP, remotePath)
  var defaultOpts = []string{
    "--progress",
    "--partial",
    "--archive",
    // "--verbose",
    "--compress",
    "--delete",
    "--keep-dirlinks",
    // "--rsh=/usr/bin/ssh",
    "--exclude=.git/",
  }
  for _, excludeOpt := range build_exclude(project_path){
    defaultOpts = append(defaultOpts, excludeOpt)
  }
  defaultOpts = append(defaultOpts, project_path)
  defaultOpts = append(defaultOpts, remote)
  cmd := exec.Command("rsync", defaultOpts...)
  // show rsync's output
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  if debug || debugoutput {
    fmt.Printf("Debug: \n%v\n", cmd.Args)
    if debug {
      return nil
    }
  }

  fmt.Printf("Sync %s.\n", project_path)
  return cmd.Run()
}
func watch(path string, ch chan<- string) {
  fmt.Printf("Watch: %s\n", path)
  var defaultOpts = []string{
    // "--print0",
    "-n",
    "--one-per-batch",
    "--recursive",
    "--exclude=.elixir_ls",
  }
  for _, excludeOpt := range build_exclude(path){
    defaultOpts = append(defaultOpts, excludeOpt)
  }
  defaultOpts = append(defaultOpts, path)
  cmd := exec.Command("fswatch", defaultOpts...)
  if debug || debugoutput {
    fmt.Printf("Debug: \n%v\n", cmd.Args)
    if debug {
      return
    }
  }
  // show rsync's output
  output, _ := cmd.StdoutPipe()
  cmd.Stderr = os.Stderr
  cmd.Start()
  scanner := bufio.NewScanner(output)
  for scanner.Scan() {
    _ = scanner.Text()
    // fmt.Println("received", m)
    ch <- path
  }
  cmd.Wait()
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
