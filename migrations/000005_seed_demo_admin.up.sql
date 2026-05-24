INSERT INTO admin_users (email, password_hash)
VALUES (
    'admin@example.com',
    '$2a$10$Hvk8p2QNmyYAvl0S3oAP0epcXG7Gjt48PdxPHAGZRmDmi3OLXRU5a'
)
ON CONFLICT (email) DO NOTHING;
