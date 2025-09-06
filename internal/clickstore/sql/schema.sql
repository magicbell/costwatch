create table if not exists %s (
  service String,
  metric String,
  value Float64,
  timestamp DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree()
TTL toDateTime(timestamp) + toIntervalDay(90)
order by (
  service,
  metric,
  timestamp
)
primary key (
  service,
  metric,
  timestamp
)
