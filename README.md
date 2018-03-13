# Kaon

Kaon is a simple URL shortening service written in Go that uses Redis for
backend storage.

## Configuration

You may set the following environment variables (default values in comments):

```bash
KAON_REDIS_HOST  # localhost
KAON_REDIS_PORT  # 6379
KAON_REDIS_DB    # 0
KAON_PORT        # 8080
```

## Usage
To shorten a URL, simply POST a form with a `url` property to the service.
You'll receive a JSON object in return, with some details about the stored link.

```bash
$ curl -XPOST -d 'url=http://example.com' http://localhost:9111/
```

```
{"Key":"1","Original":"http://example.com","Clicks":0,"CreationTime":1425315326750138320}
```

The `Key` property on the returned object defines the new short URL. You may
then nagivate to `http://localhost:9111/1` and be redirected to
[http://example.com](http://example.com).

```bash
$ curl -v http://localhost:9111/1
```

```
* Hostname was NOT found in DNS cache
*   Trying ::1...
* connect to ::1 port 9111 failed: Connection refused
*   Trying 127.0.0.1...
* Connected to localhost (127.0.0.1) port 9111 (#0)
> GET /1 HTTP/1.1
> User-Agent: curl/7.37.1
> Host: localhost:9111
> Accept: */*
>
< HTTP/1.1 301 Moved Permanently
< Location: http://example.com
< Date: Mon, 02 Mar 2015 17:05:04 GMT
< Content-Length: 53
< Content-Type: text/html; charset=utf-8
<
<a href="http://example.com">Moved Permanently</a>.
```

To view information about the stored object (such as the click count or original
URL), simply append an underscore to the `Key` value:

```
$ curl http://localhost:9111/1_
```

```
{"Key":"1","Original":"http://example.com","Clicks":3,"CreationTime":1425315326750138320}
```

# Development

Just use `docker-compose up` and you'll find kaon running on `localhost:9111`.
