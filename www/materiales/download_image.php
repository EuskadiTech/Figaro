<?php
require_once __DIR__ . "/../_incl/utils.php";
require_permission("materiales.index");

$centro = $_GET['centro'] ?? '';
$image_name = $_GET['image'] ?? '';

if (empty($centro) || empty($image_name)) {
    http_response_code(400);
    echo "Parámetros inválidos.";
    exit;
}

$image_path = "$RUTA_DATOS/Centros/$centro/Materiales/Fotos/$image_name";

if (!file_exists($image_path)) {
    http_response_code(404);
    echo "Imagen no encontrada.";
    exit;
}

$mime_type = mime_content_type($image_path);
header("Content-Type: $mime_type");
readfile($image_path);
exit;
