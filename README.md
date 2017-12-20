![hey](http://i.imgur.com/szzD9q0.png)

[Note] This is not my original work. I just made some changes according to my need. The Original work is by rakyll(https://github.com/rakyll/hey) and Tarek Ziade(https://github.com/tarekziade/boom)

hey is a tiny program that sends some load to a web application.

## Installation

    go get -u github.com/harshulsoni/hey

## Usage

hey runs provided number of requests in the provided concurrency level and prints stats.

It also supports HTTP2 endpoints.

```
Usage: hey [options...] <url>

Options:
  -n  Number of requests to run. Default is 200.
  -c1 Start range Number of requests to run concurrently. Total number of requests cannot
      be smaller than the concurrency level. Default is 50.
  -c2 End range Number of requests to run concurrently. Total number of requests cannot
      be smaller than the concurrency level. Default is 50.
  -cd Difference by which parallel connection increase in each iterations. Difference cannot be <=0 
  -q  Rate limit, in seconds (QPS).
  -o  Output type. If none provided, a summary is printed.
      "csv" is the only supported alternative. Dumps the response
      metrics in comma-separated values format.

  -m  HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
  -H  Custom HTTP header. You can specify as many as needed by repeating the flag.
      For example, -H "Accept: text/html" -H "Content-Type: application/xml" .
  -t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
  -A  HTTP Accept header.
  -d  HTTP request body.
  -D  HTTP request body from file. For example, /home/user/file.txt or ./file.txt.
  -T  Content-type, defaults to "text/html".
  -a  Basic authentication, username:password.
  -x  HTTP Proxy address as host:port.
  -h2 Enable HTTP/2.

  -host	HTTP Host header.

  -disable-compression  Disable compression.
  -disable-keepalive    Disable keep-alive, prevents re-use of TCP
                        connections between different HTTP requests.
  -cpus                 Number of used cpu cores.
                        (default for current machine is 8 cores)
```

Note: Requires go 1.7 or greater.
