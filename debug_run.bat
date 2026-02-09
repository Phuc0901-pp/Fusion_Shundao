@echo off
start /B fusion_test.exe > debug_json.txt 2>&1
timeout /t 15
curl http://localhost:5039/api/dashboard > api_response.json
taskkill /IM fusion_test.exe /F
