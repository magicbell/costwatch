create table if not exists sync_state (
  service        text not null,
  metric         text not null,
  last_synced    timestamp not null,
  last_notified  timestamp,
  primary key (service, metric)
);

create table if not exists alert_rules (
  service   text not null,
  metric    text not null,
  threshold real not null,
  primary key (service, metric)
);
