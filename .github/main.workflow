workflow "New workflow" {
  on = "push"
  resolves = ["Build"]
}

action "Build" {
  uses = "actions/docker/cli"
  runs = "docker build -t shorty:latest ."
}
