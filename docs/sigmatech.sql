-- --------------------------------------------------------
-- Host:                         localhost
-- Server version:               PostgreSQL 17.2 (Debian 17.2-1.pgdg120+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 12.2.0-14) 12.2.0, 64-bit
-- Server OS:                    
-- HeidiSQL Version:             12.1.0.6537
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES  */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

-- Dumping data for table public.customers: -1 rows
/*!40000 ALTER TABLE "customers" DISABLE KEYS */;
INSERT INTO "customers" ("uuid", "name", "email", "password", "is_active", "last_login", "created_at", "updated_at") VALUES
	('b05eea5b-e814-4128-9d8c-825544ef9e8b', 'Annisa', 'example124@gmail.com', '$2a$10$bB9iQbxLaz0nixAEnrX/RuTpiPAo0TzXjWPLzHhsrosokZfWBMQci', 'true', NULL, '2024-12-12 07:32:38.384977', '2024-12-12 07:34:26.814155'),
	('c3e5dbbb-8d31-4213-9a05-2d08873806e7', 'Budi', 'example1243@gmail.com', '$2a$10$.3.5KbOBXqHvPp5mTdbKZOirhYSsbRE0Rk/f3UlBdkC8Ca8ruBYNe', 'true', NULL, '2024-12-12 07:32:15.254986', '2024-12-12 07:35:50.69767');
/*!40000 ALTER TABLE "customers" ENABLE KEYS */;

-- Dumping data for table public.customer_information_files: -1 rows
/*!40000 ALTER TABLE "customer_information_files" DISABLE KEYS */;
INSERT INTO "customer_information_files" ("uuid", "customer_uuid", "cif_number", "nik", "full_name", "legal_name", "place_of_birth", "date_of_birth", "gender", "salary", "card_photo", "selfie_photo", "created_at", "updated_at") VALUES
	('1ffca78d-ea24-440c-8a38-f677e7ff2299', 'c3e5dbbb-8d31-4213-9a05-2d08873806e7', 'CF_000001_1733988735', '3273291115970009', 'Budi', 'Budi', 'Bandung', '1999-01-23', 'm', 10200000.00, 'customers/card-photo/05970154-25e2-4bb8-b9b7-15658a2c3498.jpg', 'customers/selfie-photo/9281d1b9-f742-4ccb-852a-5bb08d895b36.jpg', '2024-12-12 07:32:15.254986', '2024-12-12 07:32:15.254986'),
	('5ce3af46-67a8-48c2-aefa-d3266aede3c4', 'b05eea5b-e814-4128-9d8c-825544ef9e8b', 'CF_000001_1733988758', '3273291115970002', 'Annisa', 'Annisa', 'Bandung', '1999-01-23', 'm', 10200000.00, 'customers/card-photo/7b7efe2e-714d-4b4a-b006-3190c1380a88.jpg', 'customers/selfie-photo/66e43ad8-76fe-41e4-9666-979dc5aa000f.jpg', '2024-12-12 07:32:38.384977', '2024-12-12 07:32:38.384977');
/*!40000 ALTER TABLE "customer_information_files" ENABLE KEYS */;

-- Dumping data for table public.customer_limits: 8 rows
/*!40000 ALTER TABLE "customer_limits" DISABLE KEYS */;
INSERT INTO "customer_limits" ("uuid", "customer_uuid", "term", "status", "amount_limit", "remaining_limit", "created_at", "updated_at") VALUES
	('7090f00d-ef62-4551-8d63-51feaafd1974', 'b05eea5b-e814-4128-9d8c-825544ef9e8b', 1, 'true', 1000000.00, 1000000.00, '2024-12-12 07:32:38.384977', '2024-12-12 07:34:26.801787'),
	('e941c0c2-6526-4ef8-a226-796025910640', 'b05eea5b-e814-4128-9d8c-825544ef9e8b', 2, 'true', 12000000.00, 12000000.00, '2024-12-12 07:32:38.384977', '2024-12-12 07:34:26.805552'),
	('06185332-505b-442d-b36d-c49ae8ab0cf2', 'b05eea5b-e814-4128-9d8c-825544ef9e8b', 3, 'true', 1500000.00, 1500000.00, '2024-12-12 07:32:38.384977', '2024-12-12 07:34:26.808324'),
	('ea470cf1-0c3c-49a7-b389-a0e74be5e5db', 'b05eea5b-e814-4128-9d8c-825544ef9e8b', 6, 'true', 2000000.00, 2000000.00, '2024-12-12 07:32:38.384977', '2024-12-12 07:34:26.811307'),
	('600639a6-cc59-4438-bf2f-c893839c3dce', 'c3e5dbbb-8d31-4213-9a05-2d08873806e7', 1, 'true', 100000.00, 100000.00, '2024-12-12 07:32:15.254986', '2024-12-12 07:35:50.683574'),
	('695a6a45-367e-49c1-b4bc-ab5bc1005a13', 'c3e5dbbb-8d31-4213-9a05-2d08873806e7', 2, 'true', 200000.00, 200000.00, '2024-12-12 07:32:15.254986', '2024-12-12 07:35:50.686717'),
	('0a1d2a20-bdb0-4d22-8daa-6fe365f6bfca', 'c3e5dbbb-8d31-4213-9a05-2d08873806e7', 3, 'true', 500000.00, 500000.00, '2024-12-12 07:32:15.254986', '2024-12-12 07:35:50.690359'),
	('20d21c49-1429-4ee4-a338-a43bc16db7cd', 'c3e5dbbb-8d31-4213-9a05-2d08873806e7', 6, 'true', 700000.00, 700000.00, '2024-12-12 07:32:15.254986', '2024-12-12 07:35:50.693561');
/*!40000 ALTER TABLE "customer_limits" ENABLE KEYS */;

-- Dumping data for table public.goose_db_version: -1 rows
/*!40000 ALTER TABLE "goose_db_version" DISABLE KEYS */;
INSERT INTO "goose_db_version" ("id", "version_id", "is_applied", "tstamp") VALUES
	(1, 0, 'true', '2024-12-11 04:20:24.647512'),
	(2, 20241211062856, 'true', '2024-12-11 07:24:30.901078'),
	(3, 20241211063915, 'true', '2024-12-11 07:25:03.738885'),
	(4, 20241211064053, 'true', '2024-12-11 07:25:03.748829'),
	(5, 20241211064510, 'true', '2024-12-11 07:27:28.013541'),
	(6, 20241211065451, 'true', '2024-12-11 07:27:28.022329'),
	(8, 20241211072032, 'true', '2024-12-11 07:30:25.307723');
/*!40000 ALTER TABLE "goose_db_version" ENABLE KEYS */;

-- Dumping data for table public.transactions: -1 rows
/*!40000 ALTER TABLE "transactions" DISABLE KEYS */;
/*!40000 ALTER TABLE "transactions" ENABLE KEYS */;

-- Dumping data for table public.transaction_installments: -1 rows
/*!40000 ALTER TABLE "transaction_installments" DISABLE KEYS */;
/*!40000 ALTER TABLE "transaction_installments" ENABLE KEYS */;

-- Dumping data for table public.users: -1 rows
/*!40000 ALTER TABLE "users" DISABLE KEYS */;
INSERT INTO "users" ("uuid", "name", "email", "password", "last_login", "created_at", "created_by", "updated_at", "updated_by") VALUES
	('941f5f67-740f-4a5b-982e-10f42355dd81', 'Name Updated', 'super@sigmatech.id', '$2a$10$lHl1xiKdcIYpAamH.9ClwuQEbCI2Zr4mabgKLnoq0GVDAsm8gImvS', NULL, '2024-12-11 08:28:18.887823', NULL, '2024-12-11 08:28:18.887823', NULL);
/*!40000 ALTER TABLE "users" ENABLE KEYS */;

-- Dumping data for table public.variable_globals: -1 rows
/*!40000 ALTER TABLE "variable_globals" DISABLE KEYS */;
INSERT INTO "variable_globals" ("uuid", "code", "value", "description", "created_at", "updated_at") VALUES
	('7f5c7866-91b3-4115-bf28-d8cb4702f835', 'ADM', '5000', 'Admin fee @ rupiah', '2024-12-12 12:20:58.939795', '2024-12-12 12:20:58.939795'),
	('743f9557-a0c7-4c95-bf62-2db525fd691e', 'INT', '2.95', 'interest @ percentage', '2024-12-12 12:20:58.939795', '2024-12-12 12:20:58.939795');
/*!40000 ALTER TABLE "variable_globals" ENABLE KEYS */;

/*!40103 SET TIME_ZONE=IFNULL(@OLD_TIME_ZONE, 'system') */;
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;
