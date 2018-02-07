-- --------------------------------------------------------
-- Хост:                         127.0.0.1
-- Версия сервера:               10.1.13-MariaDB - mariadb.org binary distribution
-- ОС Сервера:                   Win32
-- HeidiSQL Версия:              9.3.0.4984
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;

-- Дамп структуры базы данных ticket_booking
CREATE DATABASE IF NOT EXISTS `ticket_booking` /*!40100 DEFAULT CHARACTER SET utf16 */;
USE `ticket_booking`;


-- Дамп структуры для таблица ticket_booking.events
CREATE TABLE IF NOT EXISTS `events` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf16;

-- Дамп данных таблицы ticket_booking.events: ~0 rows (приблизительно)
DELETE FROM `events`;
/*!40000 ALTER TABLE `events` DISABLE KEYS */;
INSERT INTO `events` (`id`, `name`) VALUES
	(1, 'test_event');
/*!40000 ALTER TABLE `events` ENABLE KEYS */;


-- Дамп структуры для таблица ticket_booking.event_places
CREATE TABLE IF NOT EXISTS `event_places` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `placeIdentity` varchar(255) NOT NULL DEFAULT '',
  `isBooked` int(2) DEFAULT '0',
  `isBought` int(2) DEFAULT '0',
  `userId` varchar(255) DEFAULT '0',
  `eventId` int(11) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf16;

-- Дамп данных таблицы ticket_booking.event_places: ~7 rows (приблизительно)
DELETE FROM `event_places`;
/*!40000 ALTER TABLE `event_places` DISABLE KEYS */;
INSERT INTO `event_places` (`id`, `placeIdentity`, `isBooked`, `isBought`, `userId`, `eventId`) VALUES
	(1, 'seat_1', 0, 0, '', 1),
	(2, 'seat_2', 0, 0, '', 1),
	(3, 'seat_3', 0, 0, '', 1),
	(4, 'seat_4', 0, 0, '', 1),
	(5, 'seat_5', 0, 0, '', 1),
	(6, 'seat_6', 0, 0, '', 1),
	(7, 'seat_7', 0, 0, '', 1);
/*!40000 ALTER TABLE `event_places` ENABLE KEYS */;


-- Дамп структуры для таблица ticket_booking.users
CREATE TABLE IF NOT EXISTS `users` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT '',
  `facebookId` bigint(20) unsigned DEFAULT '0',
  `email` varchar(255) DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf16;

