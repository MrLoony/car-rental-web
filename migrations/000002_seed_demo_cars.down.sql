DELETE FROM cars
WHERE slug IN (
    'toyota-corolla',
    'hyundai-elantra',
    'toyota-camry',
    'bmw-x5',
    'mercedes-benz-c-class',
    'porsche-911',
    'range-rover-sport',
    'nissan-patrol'
);

DELETE FROM car_categories
WHERE slug IN (
    'economy',
    'suv',
    'luxury',
    'sports'
);
