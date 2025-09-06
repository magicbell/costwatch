insert into alert_rules(service, metric, threshold)
values(?, ?, ?)
on conflict(service, metric) do update set threshold=excluded.threshold
