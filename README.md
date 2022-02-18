# bitbucket-2fa-monitor
A serverless function to monitor a Bitbucket workspace to ensure members are using 2FA

## Development setup

### Using Goland

1. Edit Run/Debug Configurations
1. Run kind: Directory
1. Directory: /(your path)/bitbucket-2fa-monitor/monitor
1. Output directory: /(your path)/bitbucket-2fa-monitor/bin
1. Environment:

  - API_WORKSPACE=(your workspace)
  - API_USERNAME=(your username)
  - API_APP_PASSWORD=(your app password)
  - API_BASE_URL=https://api.bitbucket.org
  - DEBUG=true
