<?php
require_once "../_incl/utils.php";
require_permission("materiales.delete");
require_once "../_incl/pre-body.php";

$centro = get_selected_centro();
$materiales_path = "$RUTA_DATOS/Centros/$centro/Materiales";
$material_id = $_GET['id'] ?? '';
$filepath = "$materiales_path/$material_id";

if (empty($material_id) || !file_exists($filepath)) {
    header("Location: index.php");
    exit();
}

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    if (isset($_POST['confirmar'])) {
        if (unlink($filepath)) {
            header("Location: index.php");
            exit();
        } else {
            $error = "No se pudo eliminar el material.";
        }
    } else {
        header("Location: index.php");
        exit();
    }
}

$material = json_decode(file_get_contents($filepath), true);
?>

<h1>Eliminar Material</h1>

<?php if (isset($error)): ?>
    <p style="color: red;"><?php echo $error; ?></p>
<?php endif; ?>

<p>¿Estás seguro de que quieres eliminar el material "<?php echo htmlspecialchars($material['nombre']); ?>"?</p>
<p>Esta acción no se puede deshacer.</p>

<form method="POST">
    <button type="submit" name="confirmar" value="1">Sí, eliminar</button>
    <a href="index.php">No, cancelar</a>
</form>

<?php
require_once "../_incl/post-body.php";
?>
