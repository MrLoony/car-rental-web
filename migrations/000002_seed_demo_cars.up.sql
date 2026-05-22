INSERT INTO car_categories (name, slug, description)
VALUES
    ('Economy', 'economy', 'Practical and affordable cars for everyday trips.'),
    ('SUV', 'suv', 'Spacious vehicles for families, groups, and longer routes.'),
    ('Luxury', 'luxury', 'Premium vehicles with added comfort and refinement.'),
    ('Sports', 'sports', 'Performance-focused cars for a more dynamic drive.')
ON CONFLICT (slug) DO UPDATE
SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO cars (
    category_id,
    brand,
    model,
    slug,
    year,
    price_per_day,
    transmission,
    fuel_type,
    seats,
    image_url,
    is_available
)
VALUES
    (
        (SELECT id FROM car_categories WHERE slug = 'economy'),
        'Toyota',
        'Corolla',
        'toyota-corolla',
        2022,
        45.00,
        'Automatic',
        'Gasoline',
        5,
        NULL,
        TRUE
    ),
    (
        (SELECT id FROM car_categories WHERE slug = 'economy'),
        'Hyundai',
        'Elantra',
        'hyundai-elantra',
        2021,
        42.00,
        'Automatic',
        'Gasoline',
        5,
        NULL,
        TRUE
    ),
    (
        (SELECT id FROM car_categories WHERE slug = 'economy'),
        'Toyota',
        'Camry',
        'toyota-camry',
        2023,
        58.00,
        'Automatic',
        'Hybrid',
        5,
        NULL,
        TRUE
    ),
    (
        (SELECT id FROM car_categories WHERE slug = 'suv'),
        'BMW',
        'X5',
        'bmw-x5',
        2022,
        120.00,
        'Automatic',
        'Diesel',
        5,
        NULL,
        TRUE
    ),
    (
        (SELECT id FROM car_categories WHERE slug = 'luxury'),
        'Mercedes-Benz',
        'C-Class',
        'mercedes-benz-c-class',
        2023,
        110.00,
        'Automatic',
        'Gasoline',
        5,
        NULL,
        TRUE
    ),
    (
        (SELECT id FROM car_categories WHERE slug = 'sports'),
        'Porsche',
        '911',
        'porsche-911',
        2021,
        240.00,
        'Automatic',
        'Gasoline',
        4,
        NULL,
        TRUE
    ),
    (
        (SELECT id FROM car_categories WHERE slug = 'suv'),
        'Range Rover',
        'Sport',
        'range-rover-sport',
        2022,
        155.00,
        'Automatic',
        'Diesel',
        5,
        NULL,
        TRUE
    ),
    (
        (SELECT id FROM car_categories WHERE slug = 'suv'),
        'Nissan',
        'Patrol',
        'nissan-patrol',
        2023,
        135.00,
        'Automatic',
        'Gasoline',
        7,
        NULL,
        TRUE
    )
ON CONFLICT (slug) DO UPDATE
SET
    category_id = EXCLUDED.category_id,
    brand = EXCLUDED.brand,
    model = EXCLUDED.model,
    year = EXCLUDED.year,
    price_per_day = EXCLUDED.price_per_day,
    transmission = EXCLUDED.transmission,
    fuel_type = EXCLUDED.fuel_type,
    seats = EXCLUDED.seats,
    image_url = EXCLUDED.image_url,
    is_available = EXCLUDED.is_available,
    updated_at = NOW();
