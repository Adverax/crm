-- Создание тестовой БД и включение pgTAP
-- Этот скрипт выполняется при первой инициализации контейнера

-- Включаем pgTAP в основной БД
CREATE EXTENSION IF NOT EXISTS pgtap;

-- Создаём тестовую БД
CREATE DATABASE crm_test OWNER crm;

-- Включаем pgTAP в тестовой БД
\c crm_test
CREATE EXTENSION IF NOT EXISTS pgtap;
