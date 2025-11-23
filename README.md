# Запуск
`docker-compose up`

Посмотреть swagger документацию и проверить работоспособность API можно, перейдя на http://localhost:8080/docs

# Допущения

Чтобы проект запускался одним docker-compose up, совсем без ручных настроек, в docker-compose.yaml указаны стандартные значения переменных окружения. Но вообще поддерживается .env файл.

Дополнительное задание "Добавить метод массовой деактивации пользователей команды и безопасную переназначаемость открытых PR (стремиться уложиться в 100 мс для средних объёмов данных)." - Под методом понимается отдельный эндпоинт, как и в первом дополнительном задании. Так как после деактивации всей команды, на открытые PR назначить некого, ревьюеры выбираются из новой команды, которую необходимо указать. Если в новой команде некого назначить, то выдаётся ошибка.

# Нагрузочное тестирование


     execution: local
        script: loadtest/loadtest.js
        output: -

     scenarios: (100.00%) 1 scenario, 5 max VUs, 1m30s max duration (incl. graceful stop):
              * default: 5 looping VUs for 1m0s (gracefulStop: 30s)

  █ THRESHOLDS

    http_req_duration
    ✓ 'p(95)<300' p(95)=12.34ms

    successful_requests
    ✓ 'rate>0.999' rate=100.00%

  █ TOTAL RESULTS

    checks_total.......: 2950    48.404892/s
    checks_succeeded...: 100.00% 2950 out of 2950
    checks_failed......: 0.00%   0 out of 2950

    ✓ health status is 200
    ✓ team/add 201 or 200/400
    ✓ team/add new 201 or 200/400
    ✓ team/get 200
    ✓ setIsActive 200 or 404
    ✓ pr/create 201 or 404/409
    ✓ users/getReview 200 or 404
    ✓ users/stats 200
    ✓ pullRequest/stats 200
    ✓ team/deactivate 200 or 404/409

    CUSTOM
    successful_requests............: 100.00% 2950 out of 2950

    HTTP
    http_req_duration..............: avg=2.99ms   min=337.62µs med=1.54ms   max=28.46ms   p(90)=7.34ms   p(95)=12.34ms
      { expected_response:true }...: avg=3.36ms   min=337.62µs med=1.66ms   max=28.46ms   p(90)=10.66ms  p(95)=12.93ms
    http_req_duration_ms...........: avg=2.997727 min=0.337627 med=1.545022 max=28.469793 p(90)=7.346589 p(95)=12.343169
    http_req_failed................: 19.93%  588 out of 2950
    http_reqs......................: 2950    48.404892/s

    EXECUTION
    iteration_duration.............: avg=1.03s    min=1.01s    med=1.03s    max=1.05s     p(90)=1.03s    p(95)=1.03s
    iterations.....................: 295     4.840489/s
    vus............................: 5       min=5            max=5
    vus_max........................: 5       min=5            max=5

    NETWORK
    data_received..................: 1.1 MB  18 kB/s
    data_sent......................: 503 kB  8.3 kB/s