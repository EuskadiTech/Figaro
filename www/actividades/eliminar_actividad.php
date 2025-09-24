<?php
require_once "../_incl/pre-body.php";

if (isset($_GET['id'])) {
    $id = $_GET['id'];
    $current_centro = $_SESSION['centro'];
    $aula = $_SESSION['aula'];
    $file_path = "$RUTA_DATOS/Actividades/$current_centro/$aula/$id.json";
    if ($_GET['global'] == '1') {
        $file_path = "$RUTA_DATOS/Actividades/_Global/$id.json";
    }
    if (file_exists($file_path)) {
        unlink($file_path);
    }
}

header("Location: index.php");
exit();
?>
