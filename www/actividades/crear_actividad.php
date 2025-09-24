<?php
require_once "../_incl/utils.php";
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Handle form submission
    $id = uniqid();
    $data = [
        'id' => $id,
        'title' => $_POST['title'],
        'start' => $_POST['start'],
        'end' => $_POST['end'],
        'description' => $_POST['description'],
        'url' => $_POST['url'],
        'meet' => $_POST['meet']
    ];

    $is_global = isset($_POST['global']) && $_POST['global'] == '1';
    $centro = $_SESSION['centro'];
    $aula = $_SESSION['aula'];
    $dir = "$RUTA_DATOS/Actividades/$centro/$aula";
    if ($is_global) {
        $dir = "$RUTA_DATOS/Actividades/_Global";
        $data["is_shared_from"] = $centro;
    }
    if (!is_dir($dir)) {
        mkdir($dir, 0777, true);
    }

    file_put_contents("$dir/$id.json", json_encode($data, JSON_PRETTY_PRINT));

    header("Location: index.php");
    exit();
}
require_once "../_incl/pre-body.php";
?>

<h2>Crear nueva actividad</h2>
<form method="POST">
    <label for="title">Título:</label><br>
    <input type="text" id="title" name="title" required><br><br>

    <label for="start">Fecha de inicio:</label><br>
    <input type="datetime-local" id="start" name="start" required><br><br>

    <label for="end">Fecha fin:</label><br>
    <input type="datetime-local" id="end" name="end" required><br><br>

    <label for="description">Descripción:</label><br>
    <textarea id="description" name="description"></textarea><br><br>

    <label for="url">Enlace:</label><br>
    <input type="url" id="url" name="url"><br><br>

    <label for="meet">Google Meet/Jitsi (embed):</label><br>
    <input type="text" id="meet" name="meet"><br><br>

    <input type="checkbox" id="global" name="global" value="1">
    <label for="global">¿Todos los centros?</label><br><br>

    <button type="submit">Crear Actividad</button>
</form>

<?php
require_once "../_incl/post-body.php";
?>
