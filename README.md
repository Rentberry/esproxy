# Elasticsearch bulk buffer
esproxy buffers and combines several ``_bulk`` requests into one and sends it after some threshold reached. Threshold may be either combined request's size or some interval.

Requests to endpoints other than ``_bulk`` will be proxied without modification.

Designed to be transparent for application - just replace your elasticsearch address with esproxy's

For example, it'll compress those 4 sequential requests 
```json
POST 127.0.0.1:19200/_bulk
{ "index" : { "_index" : "test", "_type" : "default", "_id" : "1" } }
{ "field1" : "value1" }

POST 127.0.0.1:19200/_bulk
{ "update" : { "_index" : "test", "_type" : "default", "_id" : "1" } }
{ "doc": { "field1" : "value2" } }
        
POST 127.0.0.1:19200/_bulk
{ "index" : { "_index" : "test", "_type" : "default", "_id" : "2" } }
{ "field1" : "value3" }

POST 127.0.0.1:19200/_bulk        
{ "delete" : { "_index" : "test", "_type" : "default", "_id" : "2" } }
```
into one and sends it after 20 seconds:
```json
POST elastichost:9200/_bulk
{ "index" : { "_index" : "test", "_type" : "default", "_id" : "1" } }
{ "field1" : "value1" }
{ "update" : { "_index" : "test", "_type" : "default", "_id" : "1" } }
{ "doc": { "field1" : "value2" } }
{ "index" : { "_index" : "test", "_type" : "default", "_id" : "2" } }
{ "field1" : "value3" }
{ "delete" : { "_index" : "test", "_type" : "default", "_id" : "2" } }
```

## Run
```
docker run -p 19200:19200 -p 8080:8080 \
    -e ESPROXY_ELASTICSEARCH_ADDRESS=http://elastichost:9200 \
    -e ESPROXY_FLUSH_INTERVAL=20 \
    -e ESPROXY_DEBUG=1 \
    rentberry/esproxy:latest
```

## Monitoring
esproxy exposes metrics at ``esproxy:8080/metrics``:
- `esproxy_indexer_added` - count of added records
- `esproxy_indexer_flushed` - count of flushed records
- `esproxy_indexer_failed` - count of records failed to flush
- `esproxy_indexer_indexed` - count of indexed records
- `esproxy_indexer_created` - count of created records
- `esproxy_indexer_updated` - count of updated records
- `esproxy_indexer_deleted` - count of deleted records
- `esproxy_indexer_requests` - count of requests made by indexer
- `esproxy_indexer_requests_served` - count of requests to indexer
- `esproxy_proxy_requests_served` - count of proxied requests

