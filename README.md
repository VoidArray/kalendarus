# Kalendarus Telegram Bot

* http://kalendarus.com/

## Configuration

```
log_level = "info"
calendar_url = "http://127.0.0.1:8080/basic.ics"
timezone = "Asia/Yekaterinburg"
time_format = "02.01.2006, 15:04"
pull_interval = 1800
notify_interval = 600
notify_template = "📅 %[1]s @ %[2]s\n%[3]s\n\n%[4]s"

[notification.0]
    before_start = "2h"
[notification.1]
    before_start = "24h"
[notification.2]
    before_start = "168h"

[telegram]
    enabled = true
    token = "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
    chat_id = "12345678"
    disable_web_page_preview = true
    parse_mode = "Markdown"

[plainfile]
    enabled = true
    data_dir = "/etc/kalendarus/data"
```
