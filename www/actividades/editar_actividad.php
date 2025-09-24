<?php
require_once "../_incl/pre-body.php";

$id = $_GET['id'];
$centro = $_SESSION['centro'];
$aula = $_SESSION['aula'];
$file_path = "$RUTA_DATOS/Actividades/$centro/$aula/$id.json";

if ($_GET['global'] == '1') {
    $file_path = "$RUTA_DATOS/Actividades/_Global/$id.json";
}
if (!file_exists($file_path)) {
    die("Actividad no encontrada.");
}

$data = json_decode(file_get_contents($file_path), true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Handle form submission
    $data['title'] = $_POST['title'];
    $data['start'] = $_POST['start'];
    $data['end'] = $_POST['end'];
    $data['description'] = $_POST['description'];
    $data['url'] = $_POST['url'];
    $data['meet'] = $_POST['meet'];

    $is_global = isset($_POST['global']) && $_POST['global'] == '1';
    if ($is_global) {$data["is_shared_from"] = $_SESSION['centro'];}
    $new_centro = $is_global ? '_Global' : $_SESSION['centro'];

    // If the 'global' status changes, move the file
    if ($new_centro !== $centro) {
        $new_dir = "$RUTA_DATOS/Actividades/$new_centro/$aula";
        if (!is_dir($new_dir)) {
            mkdir($new_dir, 0777, true);
        }
        $new_file_path = "$new_dir/$id.json";
        rename($file_path, $new_file_path);
        $file_path = $new_file_path;
    }

    file_put_contents($file_path, json_encode($data, JSON_PRETTY_PRINT));

    header("Location: index.php");
    exit();
}
?>

<h2>Editar actividad</h2>
<form method="POST">
    <label for="title">Título:</label><br>
    <input type="text" id="title" name="title" value="<?php echo htmlspecialchars($data['title']); ?>" required><br><br>

    <label for="start">Fecha de inicio:</label><br>
    <input type="datetime-local" id="start" name="start"
        value="<?php echo date('Y-m-d\TH:i', strtotime($data['start'])); ?>" required><br><br>

    <label for="end">Fecha de fin:</label><br>
    <input type="datetime-local" id="end" name="end"
        value="<?php echo $data['end'] ? date('Y-m-d\TH:i', strtotime($data['end'])) : ''; ?>"><br><br>

    <label for="description">Descripción:</label><br>
    <textarea id="description"
        name="description"><?php echo htmlspecialchars($data['description']); ?></textarea><br><br>

    <label for="url">Enlace:</label><br>
    <input type="url" id="url" name="url" value="<?php echo htmlspecialchars($data['url']); ?>"><br><br>

    <label for="meet">Google Meet/Jitsi (embed):</label><br>
    <input type="text" id="meet" name="meet" value="<?php echo htmlspecialchars($data['meet']); ?>"><br><br>

    <input type="checkbox" id="global" name="global" value="1" <?php echo (strpos($file_path, '_Global') !== false) ? 'checked' : ''; ?>>
    <label for="global">¿Todos los centros?</label><br><br>

    <button type="submit">Guardar Cambios</button>
</form>

<?php
require_once "../_incl/post-body.php";
?>
