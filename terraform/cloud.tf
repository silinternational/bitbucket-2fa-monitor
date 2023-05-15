terraform {
  cloud {
    organization = "gtis"

    workspaces {
      name = "bitbucket-2fa-monitor"
    }
  }
}
