# Kalendarus Telegram Bot

* http://kalendarus.com/

## Configuration

```
interval = 1800
log-level = "info"
calendar-url = "http://127.0.0.1/basic.ics"
timezone = "Asia/Yekaterinburg"
time_format = "02.01.2006, 15:04"
message_format = "%[1]s @ %[2]s\n%[3]s\n\n%[4]s"
added_event_format = "%[1]s @ %[2]s\n%[3]s\n\n%[4]s"
updated_event_format = "Перенос мероприятия %[1]s на %[2]s\n%[3]s\n\n%[4]s"
skip_first_start = true

[telegram]
    enabled = true
    token = "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
    chat-id = "12345678"

[plainfile]
    enabled = true
    data_dir = "/etc/kalendarus/data"
```
