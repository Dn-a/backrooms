# Backrooms
[![CI](https://img.shields.io/badge/CI-passed-brightgreen?logo=github)](https://github.com/Dn-a/backrooms/actions/workflows/go.yml)
[![Coverage](https://img.shields.io/badge/Coverage-44.5%25-brightgreen)](https://github.com/Dn-a/backrooms/actions/workflows/codeCoverage.yml)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.me/dnag88)

Backrooms is an easy way to redirect or reverse-proxy the client request through a single access point.

<div >
  <table><tr>
 <td style="text-align:center">
  <img width="500px"  src="assets/backrooms.png?" />
 </td>
 </tr></table>
</div>

## Simple configuration
In the same directory create a **config.yml** file, as depicted below:
```yaml
# the port on which the server is exposed
port: 4300

# Used in case a resource has no match to the endpoint called
# E.G. 
# 1. curl http://127.0.0.1:4300/test
# 2. the resource with 'matchers: /test' does not exist, 
#    then backroom performs a reverse-proxy to the default url.  
default-url: https://www.google.com

# A resource defines how each endpoint is to be managed 
resources:  
  googleR:
    name: googleR
    matchers: /**
    type: redirect    
    url: https://www.google.com

  google:
    name: google
    matchers: /search
    type: reverse-proxy
    url: https://www.google.com
```

## Issues
If you encounter problems, open an issue. Pull request are also welcome.
