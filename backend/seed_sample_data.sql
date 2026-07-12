-- AutoStock Sample Data Seed Script
-- Populates database with realistic data for a Cambodian auto garage

BEGIN;

-- ============================================================================
-- PRODUCTS
-- ============================================================================

-- Tires
INSERT INTO products (branch_id, type, sku, name, description, buy_price, sell_price, stock_quantity, min_stock_alert, tire_size, tire_brand, tire_model, tire_pattern, tire_type) VALUES
(1, 'tire', 'TIRE-MIC-20555R16', 'Michelin Primacy 4 205/55R16', 'Premium passenger tire', 80.00, 120.00, 12, 5, '205/55R16', 'Michelin', 'Primacy 4', 'Symmetric', 'passenger'),
(1, 'tire', 'TIRE-BRI-19565R15', 'Bridgestone Turanza 195/65R15', 'Comfort touring tire', 65.00, 95.00, 2, 5, '195/65R15', 'Bridgestone', 'Turanza', 'Asymmetric', 'passenger'),
(1, 'tire', 'TIRE-YOK-21560R16', 'Yokohama Advan 215/60R16', 'High performance tire', 75.00, 110.00, 6, 5, '215/60R16', 'Yokohama', 'Advan', 'Directional', 'passenger'),
(1, 'tire', 'TIRE-GOO-22550R17', 'Goodyear Eagle 225/50R17', 'Ultra high performance', 90.00, 135.00, 4, 5, '225/50R17', 'Goodyear', 'Eagle F1', 'Asymmetric', 'passenger'),
(1, 'tire', 'TIRE-CON-20560R16', 'Continental ContiSport 205/60R16', 'Sport tire', 70.00, 105.00, 10, 5, '205/60R16', 'Continental', 'ContiSport', 'Directional', 'passenger'),
(1, 'tire', 'TIRE-DUN-18565R14', 'Dunlop SP Touring 185/65R14', 'Economy tire', 45.00, 70.00, 15, 5, '185/65R14', 'Dunlop', 'SP Touring', 'Symmetric', 'passenger'),
(1, 'tire', 'TIRE-PIR-23545R18', 'Pirelli P Zero 235/45R18', 'Ultra high performance', 110.00, 165.00, 0, 5, '235/45R18', 'Pirelli', 'P Zero', 'Asymmetric', 'passenger'),
(1, 'tire', 'TIRE-HAN-21555R17', 'Hankook Ventus 215/55R17', 'Performance tire', 68.00, 100.00, 7, 5, '215/55R17', 'Hankook', 'Ventus V12', 'Directional', 'passenger'),
(1, 'tire', 'TIRE-BRI-24540R18', 'Bridgestone Potenza 245/40R18', 'Ultra high performance sport tire', 120.00, 180.00, 3, 5, '245/40R18', 'Bridgestone', 'Potenza S001', 'Asymmetric', 'passenger'),
(1, 'tire', 'TIRE-YOK-17570R14', 'Yokohama BluEarth 175/70R14', 'Economy eco tire', 40.00, 65.00, 8, 5, '175/70R14', 'Yokohama', 'BluEarth', 'Symmetric', 'passenger'),
(1, 'tire', 'TIRE-MIC-22565R17', 'Michelin Latitude 225/65R17', 'SUV tire', 95.00, 145.00, 5, 5, '225/65R17', 'Michelin', 'Latitude Sport', 'Asymmetric', 'suv'),
(1, 'tire', 'TIRE-CON-23555R19', 'Continental CrossContact 235/55R19', 'SUV all-season tire', 100.00, 155.00, 4, 5, '235/55R19', 'Continental', 'CrossContact', 'Directional', 'suv'),
(1, 'tire', 'TIRE-GOO-26570R17', 'Goodyear Wrangler 265/70R17', 'Truck/SUV tire', 105.00, 160.00, 6, 5, '265/70R17', 'Goodyear', 'Wrangler AT', 'All-Terrain', 'truck'),
(1, 'tire', 'TIRE-DUN-19555R16', 'Dunlop Sport 195/55R16', 'Sport compact tire', 55.00, 85.00, 9, 5, '195/55R16', 'Dunlop', 'Sport SP', 'Directional', 'passenger');

-- Parts
INSERT INTO products (branch_id, type, sku, name, description, buy_price, sell_price, stock_quantity, min_stock_alert) VALUES
(1, 'part', 'PART-OIL-TOY-001', 'Oil Filter Toyota', 'Genuine Toyota oil filter', 5.00, 12.00, 25, 10),
(1, 'part', 'PART-OIL-HON-001', 'Oil Filter Honda', 'Genuine Honda oil filter', 5.00, 12.00, 3, 10),
(1, 'part', 'PART-BRK-TOY-001', 'Brake Pads Front Toyota', 'Front brake pads set for Toyota', 25.00, 55.00, 8, 5),
(1, 'part', 'PART-BRK-HON-001', 'Brake Pads Front Honda', 'Front brake pads set for Honda', 25.00, 55.00, 6, 5),
(1, 'part', 'PART-SPK-UNI-001', 'Spark Plugs Iridium Set', 'Iridium spark plugs set of 4', 20.00, 45.00, 15, 10),
(1, 'part', 'PART-AIR-TOY-001', 'Air Filter Toyota', 'Engine air filter for Toyota', 8.00, 18.00, 12, 8),
(1, 'part', 'PART-AIR-HON-001', 'Air Filter Honda', 'Engine air filter for Honda', 8.00, 18.00, 2, 8),
(1, 'part', 'PART-BAT-UNI-001', 'Car Battery 12V NS60', 'Maintenance-free 12V battery', 60.00, 120.00, 5, 3),
(1, 'part', 'PART-WIP-UNI-001', 'Wiper Blades Universal', 'Universal wiper blades (pair)', 8.00, 18.00, 20, 10),
(1, 'part', 'PART-BULB-UNI-001', 'Headlight Bulb H4', 'Halogen headlight bulb', 6.00, 15.00, 18, 10),
(1, 'part', 'PART-BELT-UNI-001', 'Serpentine Belt', 'Engine drive belt universal', 18.00, 38.00, 7, 5),
(1, 'part', 'PART-ALT-TOY-001', 'Alternator Toyota Camry', 'Alternator for Toyota Camry 2018-2022', 85.00, 175.00, 0, 2),
(1, 'part', 'PART-HOS-UNI-001', 'Upper Radiator Hose', 'Universal radiator hose', 12.00, 28.00, 11, 5),
(1, 'part', 'PART-CAL-TOY-001', 'Brake Caliper Left Front Toyota', 'Front left brake caliper', 35.00, 75.00, 4, 3);

-- Labor
INSERT INTO products (branch_id, type, sku, name, description, buy_price, sell_price, stock_quantity, min_stock_alert, unit) VALUES
(1, 'labor', 'LABOR-TIRE-INSTALL', 'Tire Installation', 'Mount and balance one tire', 0.00, 15.00, 999, 0, 'piece'),
(1, 'labor', 'LABOR-WHEEL-ALIGN', 'Wheel Alignment', 'Four-wheel alignment', 0.00, 35.00, 999, 0, 'service'),
(1, 'labor', 'LABOR-OIL-CHANGE', 'Oil Change Service', 'Engine oil and filter change', 0.00, 25.00, 999, 0, 'service'),
(1, 'labor', 'LABOR-BRAKE-SERVICE', 'Brake Service', 'Front brake pads replacement', 0.00, 45.00, 999, 0, 'service'),
(1, 'labor', 'LABOR-TUNE-UP', 'Engine Tune-Up', 'Spark plugs replacement and basic tune-up', 0.00, 65.00, 999, 0, 'service'),
(1, 'labor', 'LABOR-INSPECTION', 'Multi-Point Inspection', 'Vehicle inspection service', 0.00, 20.00, 999, 0, 'service'),
(1, 'labor', 'LABOR-BATTERY-INSTALL', 'Battery Installation', 'Battery replacement service', 0.00, 10.00, 999, 0, 'service'),
(1, 'labor', 'LABOR-DIAGNOSTIC', 'Diagnostic Check', 'Computer diagnostic scan', 0.00, 30.00, 999, 0, 'service');

-- Consumables
INSERT INTO products (branch_id, type, sku, name, description, buy_price, sell_price, stock_quantity, min_stock_alert, unit) VALUES
(1, 'consumable', 'CONS-OIL-5W30', 'Engine Oil 5W-30', 'Synthetic engine oil 1 liter', 8.00, 15.00, 30, 15, 'liter'),
(1, 'consumable', 'CONS-OIL-10W40', 'Engine Oil 10W-40', 'Semi-synthetic engine oil 1 liter', 6.00, 12.00, 4, 15, 'liter'),
(1, 'consumable', 'CONS-OIL-0W20', 'Engine Oil 0W-20', 'Full synthetic engine oil 1 liter', 9.00, 18.00, 18, 10, 'liter'),
(1, 'consumable', 'CONS-BRK-FLUID', 'Brake Fluid DOT 4', 'Brake fluid 500ml bottle', 5.00, 12.00, 15, 8, 'bottle'),
(1, 'consumable', 'CONS-COOLANT', 'Engine Coolant Antifreeze', 'Engine coolant 1 liter', 4.00, 10.00, 20, 10, 'liter'),
(1, 'consumable', 'CONS-WASH-FLUID', 'Windshield Washer Fluid', 'Washer fluid 1 liter', 2.00, 5.00, 30, 15, 'liter'),
(1, 'consumable', 'CONS-POWER-STEER', 'Power Steering Fluid', 'Power steering fluid 500ml', 5.00, 11.00, 8, 5, 'bottle'),
(1, 'consumable', 'CONS-TRANS-FLUID', 'Transmission Fluid ATF', 'Automatic transmission fluid 1 liter', 10.00, 22.00, 6, 5, 'liter');

-- ============================================================================
-- CUSTOMERS
-- ============================================================================

INSERT INTO customers (branch_id, name, phone, email, address, notes, customer_since) VALUES
(1, 'Sokha Phan', '+855 12 345 678', 'sokha.phan@email.com', 'Street 63, Phnom Penh', 'Regular customer, owns Toyota Camry and Honda Civic', '2024-01-15'),
(1, 'Vannak Chea', '+855 17 234 567', 'vannak.chea@email.com', 'Street 310, Phnom Penh', 'Business customer with fleet of 3 vehicles', '2024-02-20'),
(1, 'Dara Kim', '+855 92 123 456', 'dara.kim@email.com', 'Siem Reap Province', 'Honda Civic owner, seasonal visitor', '2024-03-10'),
(1, 'Srey Mom', '+855 16 789 012', 'srey.mom@email.com', 'Street 128, Phnom Penh', 'New customer, recommended by friend', '2024-06-01'),
(1, 'Pisey Lim', '+855 93 456 789', 'pisey.lim@email.com', 'Battambang Province', 'Suzuki Swift owner', '2024-04-15'),
(1, 'Rithy Sok', '+855 11 234 567', 'rithy.sok@email.com', 'Street 200, Phnom Penh', 'Long-time customer, owns multiple vehicles', '2023-11-20'),
(1, 'Channa Mao', '+855 96 345 678', 'channa.mao@email.com', 'Street 51, Phnom Penh', 'Toyota Vios owner', '2024-05-10'),
(1, 'Sovannara Tep', '+855 15 678 901', 'sovannara.tep@email.com', 'Kampong Cham Province', 'Honda Accord owner', '2024-01-25'),
(1, 'Bopha Nguon', '+855 99 012 345', 'bopha.nguon@email.com', 'Street 271, Phnom Penh', 'New customer', '2024-06-15'),
(1, 'Kosal Phan', '+855 12 987 654', 'kosal.phan@email.com', 'Street 110, Phnom Penh', 'Business customer, runs delivery service', '2024-02-05'),
(1, 'Sreynich Heng', '+855 17 654 321', 'sreynich.heng@email.com', 'Siem Reap Province', 'Regular customer when visiting Phnom Penh', '2024-03-20'),
(1, 'Visal Chan', '+855 92 987 654', 'visal.chan@email.com', 'Street 139, Phnom Penh', 'Toyota Corolla owner', '2024-04-01'),
(1, 'Moneath Lay', '+855 16 543 210', 'moneath.lay@email.com', 'Street 450, Phnom Penh', 'Regular tire customer', '2023-12-01'),
(1, 'Sophea Thong', '+855 93 210 987', 'sophea.thong@email.com', 'Kandal Province', 'New SUV owner', '2024-07-01'),
(1, 'Rathana Chey', '+855 11 876 543', 'rathana.chey@email.com', 'Street 123, Phnom Penh', 'Premium tire customer', '2024-05-20');

-- ============================================================================
-- VEHICLES
-- ============================================================================

INSERT INTO vehicles (customer_id, plate_number, make, model, year, vin, color, notes) VALUES
(1, 'PP 2A-1234', 'Toyota', 'Camry', 2020, '1HGCM82633A123456', 'Silver', 'Primary vehicle, regular maintenance'),
(1, 'PP 2A-5678', 'Honda', 'Civic', 2019, '2HGFC2F59KH123456', 'Black', 'Secondary vehicle, used for weekend trips'),
(2, 'PP 2B-3456', 'Toyota', 'Vios', 2022, 'JTDKN3DU5A1234567', 'White', 'Fleet vehicle 1 - used daily'),
(2, 'PP 2B-7890', 'Toyota', 'Hilux', 2021, '5TFDW5F17LX123456', 'Red', 'Fleet vehicle 2 - cargo deliveries'),
(3, 'SR 3C-2345', 'Honda', 'Civic', 2018, '2HGFC2F58JH123456', 'Blue', 'Siem Reap registered'),
(4, 'PP 4D-6789', 'Suzuki', 'Swift', 2022, 'JSA3AY82S00123456', 'Gray', 'Recently purchased'),
(5, 'BT 5E-1357', 'Suzuki', 'Swift', 2020, 'JSA3AY82S00234567', 'White', 'Battambang registered'),
(6, 'PP 6F-2468', 'Toyota', 'Camry', 2021, '1HGCM82633A234567', 'Black', 'Primary vehicle'),
(6, 'PP 6F-9753', 'Honda', 'Accord', 2020, '1HGCV1F30LA123456', 'Silver', 'Secondary vehicle'),
(7, 'PP 7G-4680', 'Toyota', 'Vios', 2022, 'JTDKN3DU5A2345678', 'White', 'Daily commute'),
(8, 'KC 8H-1359', 'Honda', 'Accord', 2021, '1HGCV1F30MA234567', 'Gray', 'Kampong Cham registered'),
(9, 'PP 9J-2460', 'Toyota', 'Corolla', 2020, '5YFEPRAE0LP123456', 'Blue', ''),
(10, 'PP 10K-3579', 'Toyota', 'Hilux', 2021, '5TFDW5F17MX234567', 'Black', 'Business use - delivery vehicle'),
(11, 'SR 11L-4680', 'Honda', 'Civic', 2019, '2HGFC2F59KH234567', 'Red', 'Siem Reap registered'),
(12, 'PP 12M-5791', 'Toyota', 'Corolla', 2021, '5YFEPRAE0MP234567', 'Silver', ''),
(13, 'PP 13N-6802', 'Toyota', 'Camry', 2018, '1HGCM82633A345678', 'White', 'Previous model, still runs well'),
(14, 'KD 14P-7913', 'Toyota', 'Fortuner', 2023, '5TDMW5F1XPX123456', 'Black', 'New SUV, custom service'),
(15, 'PP 15R-8024', 'Lexus', 'ES 350', 2022, 'JTHB61B51N1234567', 'Silver', 'Premium customer vehicle');

-- ============================================================================
-- SERVICE JOBS
-- ============================================================================

-- Completed jobs (Day -6 to Day -1)
INSERT INTO service_jobs (branch_id, customer_id, vehicle_id, job_number, status, priority, description, diagnosis, work_performed, estimated_hours, actual_hours, started_at, completed_at, notes) VALUES
(1, 1, 1, 'JOB-2026-0001', 'completed', 'normal', 'Tire replacement and alignment', 'Front tires worn unevenly, alignment off by 2 degrees', 'Replaced 2 front Michelin tires, performed 4-wheel alignment', 2.0, 1.5, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days' + INTERVAL '1 hour 30 minutes', 'Customer preferred Michelin tires'),
(1, 2, 3, 'JOB-2026-0002', 'completed', 'high', 'Oil change and multi-point inspection', 'Regular fleet maintenance', 'Changed oil and filter, inspected all systems, topped off fluids', 1.0, 0.75, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days' + INTERVAL '45 minutes', 'Fleet vehicle - priority service'),
(1, 3, 5, 'JOB-2026-0003', 'completed', 'normal', 'Front brake pad replacement', 'Squeaking noise from front brakes, pads at 3mm', 'Replaced front brake pads, resurfaced rotors, tested', 1.5, 1.25, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days' + INTERVAL '1 hour 15 minutes', ''),
(1, 4, 6, 'JOB-2026-0004', 'completed', 'normal', 'Tire rotation and balancing', 'Regular tire maintenance', 'Rotated all 4 tires, balanced and adjusted pressure', 0.5, 0.5, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days' + INTERVAL '30 minutes', ''),
(1, 5, 7, 'JOB-2026-0005', 'completed', 'low', 'Spark plug replacement and tune-up', 'Engine misfiring at idle, rough running', 'Replaced spark plugs with iridium set, performed full tune-up', 2.0, 1.75, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days' + INTERVAL '1 hour 45 minutes', ''),
(1, 6, 8, 'JOB-2026-0006', 'completed', 'high', 'Full brake service both axles', 'Brake pedal feels soft, stopping distance increased', 'Replaced all brake pads, resurfaced front rotors, bled brake system', 3.0, 2.5, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '2 hours 30 minutes', ''),
(1, 7, 10, 'JOB-2026-0007', 'completed', 'normal', 'Battery replacement', 'Battery not holding charge, 3 years old', 'Tested battery (failed), replaced with new NS60 battery, tested charging system', 0.5, 0.5, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '30 minutes', ''),
(1, 9, 12, 'JOB-2026-0008', 'completed', 'normal', 'Air filter and oil change', 'Regular service due', 'Changed engine oil, replaced oil filter and air filter', 1.0, 0.75, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '45 minutes', ''),
(1, 10, 13, 'JOB-2026-0009', 'completed', 'normal', 'Wheel alignment and tire inspection', 'Vehicle pulling to left while driving', 'Performed 4-wheel alignment, inspected and rotated tires', 1.5, 1.0, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '1 hour', 'Alignment corrected');

-- In progress jobs
INSERT INTO service_jobs (branch_id, customer_id, vehicle_id, job_number, status, priority, description, diagnosis, work_performed, estimated_hours, actual_hours, started_at, completed_at, notes) VALUES
(1, 11, 14, 'JOB-2026-0010', 'in_progress', 'high', 'Brake and transmission service', 'Brake fluid dirty, transmission shifting rough', 'Started brake service, transmission flush in progress', 3.5, NULL, NOW() - INTERVAL '1 day', NULL, 'Waiting for transmission fluid delivery'),
(1, 12, 15, 'JOB-2026-0011', 'in_progress', 'normal', 'Routine maintenance - oil, filters, inspection', 'Regular 6-month maintenance', 'Oil change completed, working through inspection checklist', 1.5, 0.5, NOW(), NULL, '');

-- Pending jobs
INSERT INTO service_jobs (branch_id, customer_id, vehicle_id, job_number, status, priority, description, diagnosis, work_performed, estimated_hours, actual_hours, started_at, completed_at, notes) VALUES
(1, 8, 11, 'JOB-2026-0012', 'pending', 'normal', 'Wheel alignment and balancing', 'Vehicle vibrates at highway speed', NULL, 1.0, NULL, NULL, NULL, 'Scheduled for later today'),
(1, 13, 16, 'JOB-2026-0013', 'pending', 'normal', 'New tire purchase and installation', '4 tires worn out, needs replacement', NULL, 1.5, NULL, NULL, NULL, 'Waiting for Michelin tire delivery'),
(1, 14, 17, 'JOB-2026-0014', 'pending', 'urgent', 'A/C system not cooling', 'Air conditioning blowing hot air', NULL, 2.0, NULL, NULL, NULL, 'Customer called ahead - urgent'),
(1, 15, 18, 'JOB-2026-0015', 'pending', 'high', 'Diagnostic check - check engine light', 'Check engine light on since yesterday', NULL, 1.0, NULL, NULL, NULL, 'Lexus customer - priority');

-- ============================================================================
-- SERVICE JOB ITEMS
-- ============================================================================

-- Job 1 (Tire replacement and alignment)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(1, 1, 'Michelin Primacy 4 205/55R16', 2, 120.00, 240.00),
(1, 29, 'Tire Installation', 2, 15.00, 30.00),
(1, 30, 'Wheel Alignment', 1, 35.00, 35.00);

-- Job 2 (Oil change and inspection)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(2, 15, 'Oil Filter Toyota', 1, 12.00, 12.00),
(2, 33, 'Engine Oil 5W-30', 4, 15.00, 60.00),
(2, 31, 'Oil Change Service', 1, 25.00, 25.00),
(2, 34, 'Multi-Point Inspection', 1, 20.00, 20.00);

-- Job 3 (Brake service)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(3, 17, 'Brake Pads Front Honda', 1, 55.00, 55.00),
(3, 32, 'Brake Service', 1, 45.00, 45.00);

-- Job 4 (Tire rotation)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(4, 29, 'Tire Installation', 4, 15.00, 60.00);

-- Job 5 (Tune-up)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(5, 19, 'Spark Plugs Iridium Set', 1, 45.00, 45.00),
(5, 33, 'Engine Tune-Up', 1, 65.00, 65.00);

-- Job 6 (Full brake service)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(6, 17, 'Brake Pads Front Toyota', 1, 55.00, 55.00),
(6, 18, 'Brake Pads Front Honda', 1, 55.00, 55.00),
(6, 32, 'Brake Service', 2, 45.00, 90.00);

-- Job 7 (Battery replacement)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(7, 22, 'Car Battery 12V NS60', 1, 120.00, 120.00),
(7, 35, 'Battery Installation', 1, 10.00, 10.00);

-- Job 8 (Oil + air filter change)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(8, 15, 'Oil Filter Toyota', 1, 12.00, 12.00),
(8, 33, 'Engine Oil 5W-30', 4, 15.00, 60.00),
(8, 20, 'Air Filter Toyota', 1, 18.00, 18.00),
(8, 31, 'Oil Change Service', 1, 25.00, 25.00);

-- Job 9 (Alignment and tire inspection)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(9, 30, 'Wheel Alignment', 1, 35.00, 35.00),
(9, 29, 'Tire Installation', 4, 15.00, 60.00);

-- Job 10 (Brake + transmission - in progress)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(10, 18, 'Brake Pads Front Honda', 1, 55.00, 55.00),
(10, 32, 'Brake Service', 1, 45.00, 45.00),
(10, 40, 'Transmission Fluid ATF', 4, 22.00, 88.00);

-- Job 11 (Routine maintenance - in progress)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(11, 15, 'Oil Filter Toyota', 1, 12.00, 12.00),
(11, 33, 'Engine Oil 5W-30', 4, 15.00, 60.00),
(11, 21, 'Air Filter Honda', 1, 18.00, 18.00),
(11, 34, 'Multi-Point Inspection', 1, 20.00, 20.00);

-- Job 13 (Pending - tire purchase)
INSERT INTO service_job_items (service_job_id, product_id, description, quantity, unit_price, total_price) VALUES
(13, 1, 'Michelin Primacy 4 205/55R16', 4, 120.00, 480.00),
(13, 29, 'Tire Installation', 4, 15.00, 60.00),
(13, 30, 'Wheel Alignment', 1, 35.00, 35.00);

-- ============================================================================
-- INVOICES
-- ============================================================================

-- Paid invoices (last 7 days)
INSERT INTO invoices (branch_id, invoice_number, customer_id, vehicle_id, service_job_id, subtotal, tax_rate, tax_amount, discount, total_usd, exchange_rate, total_khr, payment_status, status, paid_amount, payment_method, issued_at, created_at) VALUES
(1, 'INV-2026-0001', 1, 1, 1, 305.00, 0, 0, 0, 305.00, 4050, 1235250, 'paid', 'paid', 305.00, 'cash', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
(1, 'INV-2026-0002', 2, 3, 2, 117.00, 0, 0, 0, 117.00, 4050, 473850, 'paid', 'paid', 117.00, 'transfer', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(1, 'INV-2026-0003', 3, 5, 3, 100.00, 0, 0, 0, 100.00, 4050, 405000, 'paid', 'paid', 100.00, 'cash', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(1, 'INV-2026-0004', 4, 6, 4, 60.00, 0, 0, 0, 60.00, 4050, 243000, 'paid', 'paid', 60.00, 'card', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(1, 'INV-2026-0005', 5, 7, 5, 110.00, 0, 0, 10, 100.00, 4050, 405000, 'paid', 'paid', 100.00, 'cash', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
(1, 'INV-2026-0006', 6, 8, 6, 200.00, 0, 0, 0, 200.00, 4050, 810000, 'paid', 'paid', 200.00, 'card', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(1, 'INV-2026-0007', 7, 10, 7, 130.00, 0, 0, 0, 130.00, 4050, 526500, 'paid', 'paid', 130.00, 'cash', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(1, 'INV-2026-0008', 9, 12, 8, 115.00, 0, 0, 0, 115.00, 4050, 465750, 'paid', 'paid', 115.00, 'transfer', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
(1, 'INV-2026-0009', 10, 13, 9, 95.00, 0, 0, 0, 95.00, 4050, 384750, 'paid', 'paid', 95.00, 'cash', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

-- Issued invoices (unpaid or partial)
INSERT INTO invoices (branch_id, invoice_number, customer_id, vehicle_id, service_job_id, subtotal, tax_rate, tax_amount, discount, total_usd, exchange_rate, total_khr, payment_status, status, paid_amount, payment_method, payment_notes, issued_at, created_at) VALUES
(1, 'INV-2026-0010', 11, 14, 10, 188.00, 0, 0, 0, 188.00, 4050, 761400, 'unpaid', 'issued', 0, NULL, NULL, NOW(), NOW()),
(1, 'INV-2026-0011', 12, 15, 11, 110.00, 0, 0, 0, 110.00, 4050, 445500, 'partial', 'issued', 50.00, 'cash', 'Partial payment $50', NOW(), NOW());

-- Walk-in invoice (no service job)
INSERT INTO invoices (branch_id, invoice_number, customer_id, vehicle_id, subtotal, tax_rate, tax_amount, discount, total_usd, exchange_rate, total_khr, payment_status, status, paid_amount, payment_method, issued_at, created_at) VALUES
(1, 'INV-2026-0012', 1, 1, 240.00, 0, 0, 0, 240.00, 4050, 972000, 'paid', 'paid', 240.00, 'card', NOW(), NOW());

-- Voided invoice
INSERT INTO invoices (branch_id, invoice_number, customer_id, vehicle_id, subtotal, tax_rate, tax_amount, discount, total_usd, exchange_rate, total_khr, payment_status, status, voided_at, void_reason, voided_by, issued_at, created_at) VALUES
(1, 'INV-2026-0013', 8, 11, 120.00, 0, 0, 0, 120.00, 4050, 486000, 'voided', 'voided', NOW() - INTERVAL '1 day', 'Customer cancelled after ordering wrong tire size', 1, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days');

-- ============================================================================
-- INVOICE ITEMS
-- ============================================================================

-- INV-2026-0001
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(1, 1, 'product', 'Michelin Primacy 4 205/55R16', 2, 120.00, 240.00),
(1, 29, 'labor', 'Tire Installation', 2, 15.00, 30.00),
(1, 30, 'labor', 'Wheel Alignment', 1, 35.00, 35.00);

-- INV-2026-0002
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(2, 15, 'product', 'Oil Filter Toyota', 1, 12.00, 12.00),
(2, 33, 'product', 'Engine Oil 5W-30', 4, 15.00, 60.00),
(2, 31, 'labor', 'Oil Change Service', 1, 25.00, 25.00),
(2, 34, 'labor', 'Multi-Point Inspection', 1, 20.00, 20.00);

-- INV-2026-0003
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(3, 18, 'product', 'Brake Pads Front Honda', 1, 55.00, 55.00),
(3, 32, 'labor', 'Brake Service', 1, 45.00, 45.00);

-- INV-2026-0004
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(4, 29, 'labor', 'Tire Installation', 4, 15.00, 60.00);

-- INV-2026-0005
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(5, 19, 'product', 'Spark Plugs Iridium Set', 1, 45.00, 45.00),
(5, 33, 'labor', 'Engine Tune-Up', 1, 65.00, 65.00);

-- INV-2026-0006
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(6, 17, 'product', 'Brake Pads Front Toyota', 1, 55.00, 55.00),
(6, 18, 'product', 'Brake Pads Front Honda', 1, 55.00, 55.00),
(6, 32, 'labor', 'Brake Service', 2, 45.00, 90.00);

-- INV-2026-0007
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(7, 22, 'product', 'Car Battery 12V NS60', 1, 120.00, 120.00),
(7, 35, 'labor', 'Battery Installation', 1, 10.00, 10.00);

-- INV-2026-0008
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(8, 15, 'product', 'Oil Filter Toyota', 1, 12.00, 12.00),
(8, 33, 'product', 'Engine Oil 5W-30', 4, 15.00, 60.00),
(8, 20, 'product', 'Air Filter Toyota', 1, 18.00, 18.00),
(8, 31, 'labor', 'Oil Change Service', 1, 25.00, 25.00);

-- INV-2026-0009
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(9, 30, 'labor', 'Wheel Alignment', 1, 35.00, 35.00),
(9, 29, 'labor', 'Tire Installation', 4, 15.00, 60.00);

-- INV-2026-0010 (unpaid)
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(10, 18, 'product', 'Brake Pads Front Honda', 1, 55.00, 55.00),
(10, 32, 'labor', 'Brake Service', 1, 45.00, 45.00),
(10, 40, 'product', 'Transmission Fluid ATF', 4, 22.00, 88.00);

-- INV-2026-0011 (partial)
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(11, 15, 'product', 'Oil Filter Toyota', 1, 12.00, 12.00),
(11, 33, 'product', 'Engine Oil 5W-30', 4, 15.00, 60.00),
(11, 21, 'product', 'Air Filter Honda', 1, 18.00, 18.00),
(11, 34, 'labor', 'Multi-Point Inspection', 1, 20.00, 20.00);

-- INV-2026-0012 (walk-in)
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(12, 2, 'product', 'Bridgestone Turanza 195/65R15', 2, 95.00, 190.00),
(12, 29, 'labor', 'Tire Installation', 2, 15.00, 30.00),
(12, 30, 'labor', 'Wheel Alignment', 1, 35.00, 35.00);

-- INV-2026-0013 (voided)
INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd) VALUES
(13, 1, 'product', 'Michelin Primacy 4 205/55R16', 1, 120.00, 120.00);

COMMIT;

-- Products start with no image_url (ProductThumb renders a tinted type-icon
-- fallback until a real photo is uploaded). Real photos are applied per
-- product via the upload endpoint (POST /products/:id/image) — for a demo
-- seed, one representative photo per type (tire/part/labor/consumable) is
-- enough; see the deploy notes for the one-off script that applied this to
-- the initial 44 seeded products.
