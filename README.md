Feedly exporter for Prometheus
===============================

Get the subscription data for your blog.

Metrics:

* feedly\_subscribers - count of subscriptions
* feedly\_response - remote query duration
* feedly\_velocity - publishing schedule

Example:

Run the exporter:

```
$ go run app.go -urls=http://blog.alexellis.io/rss/,http://jmkhael.io/rss/,https://finnian.io/blog/rss/,http://blog.technologee.co.uk/rss/,http://rbkr.ddns.net/rss/,http://brianchristner.io/rss/
```

Now use `curl` or add it to your scrape_config section:

```
echo ; curl -s localhost:9001/metrics|grep feedly; echo

# HELP feedly_response remote response duration
# TYPE feedly_response gauge
feedly_response{url="http://blog.alexellis.io/rss/"} 0.465507394
feedly_response{url="http://jmkhael.io/rss/"} 1.585638863
# HELP feedly_subscribers count of subscriptions
# TYPE feedly_subscribers gauge
feedly_subscribers{url="http://blog.alexellis.io/rss/"} 330
feedly_subscribers{url="http://jmkhael.io/rss/"} 10
# HELP feedly_velocity velocity of publishing
# TYPE feedly_velocity gauge
feedly_velocity{url="http://blog.alexellis.io/rss/"} 1.4
feedly_velocity{url="http://jmkhael.io/rss/"} 0.2
```

License is MIT.

