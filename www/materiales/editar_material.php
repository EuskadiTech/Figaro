<?php
require_once "../_incl/utils.php";
require_permission("materiales.update");
require_once "../_incl/pre-body.php";

$centro = get_selected_centro();
$materiales_path = "$RUTA_DATOS/Centros/$centro/Materiales";
$material_id = $_GET['id'] ?? '';
$filepath = "$materiales_path/$material_id";

if (empty($material_id) || !file_exists($filepath)) {
    header("Location: index.php");
    exit();
}

$material = json_decode(file_get_contents($filepath), true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $nombre = $_POST['nombre'] ?? $material['nombre'];
    $unidad = $_POST['unidad'] ?? $material['unidad'];
    $cantidad_disponible = $_POST['cantidad_disponible'] ?? $material['cantidad_disponible'];
    $cantidad_minima = $_POST['cantidad_minima'] ?? $material['cantidad_minima'];
    $notas = $_POST['notas'] ?? $material['notas'];
    $foto_path = $material['foto'];

    $uploads_dir = "$RUTA_DATOS/Centros/$centro/Materiales/Fotos";
    if (isset($_FILES['foto']) && $_FILES['foto']['error'] == UPLOAD_ERR_OK) {
        $tmp_name = $_FILES['foto']['tmp_name'];
        $img_name = uniqid('img_') . '_' . basename($_FILES['foto']['name']);
        $target_file = $uploads_dir . '/' . $img_name;
        if (move_uploaded_file($tmp_name, $target_file)) {
            $foto_path = "download_image.php?centro=$centro&image=$img_name";
        }
    }

    $data = [
        'nombre' => $nombre,
        'foto' => $foto_path,
        'unidad' => $unidad,
        'cantidad_disponible' => $cantidad_disponible,
        'cantidad_minima' => $cantidad_minima,
        'notas' => $notas,
        'createdAt' => $material['createdAt'],
        'updatedAt' => date('c')
    ];

    if (file_put_contents($filepath, json_encode($data, JSON_PRETTY_PRINT))) {
        header("Location: index.php");
        exit();
    } else {
        $error = "No se pudo guardar el material.";
    }
}

?>

<h1>Editar Material</h1>

<?php if (isset($error)): ?>
    <p style="color: red;"><?php echo $error; ?></p>
<?php endif; ?>

<form method="POST" enctype="multipart/form-data">
    <fieldset>
        <legend>Editando: <?php echo htmlspecialchars($material['nombre']); ?></legend>

        <label for="nombre">Nombre del material:</label><br>
        <input type="text" id="nombre" name="nombre" value="<?php echo htmlspecialchars($material['nombre']); ?>" required><br><br>

        <label for="foto">Foto:<br>
        <img id="fotoPreview" src="<?php echo htmlspecialchars($material['foto']); ?>" style="max-height: 200px; max-width: 100%; border: 1px solid #ccc; padding: 10px; border-radius: 5px;"><br>
        </label>
        <input style="display: none;" type="file" id="foto" name="foto" accept="image/*"><br><br>
        
        <script>
            const fotoInput = document.getElementById('foto');
            const fotoPreview = document.getElementById('fotoPreview');
            const fotoLabel = document.querySelector('label[for="foto"]');

            fotoLabel.addEventListener('click', () => fotoInput.click());

            fotoInput.addEventListener('change', function() {
                const file = this.files[0];
                if (file) {
                    const reader = new FileReader();
                    reader.onload = function(e) {
                        fotoPreview.src = e.target.result;
                    }
                    reader.readAsDataURL(file);
                }
            });
        </script>

        <label for="unidad">Unidad de medida:</label><br>
        <select id="unidad" name="unidad">
            <option value="unidad" <?php if ($material['unidad'] == 'unidad') echo 'selected'; ?>>Unidad</option>
            <option value="caja" <?php if ($material['unidad'] == 'caja') echo 'selected'; ?>>Caja</option>
            <option value="paquete" <?php if ($material['unidad'] == 'paquete') echo 'selected'; ?>>Paquete</option>
            <option value="docena" <?php if ($material['unidad'] == 'docena') echo 'selected'; ?>>Docena</option>
            <optgroup label="Metrico">
                <option value="gramos" <?php if ($material['unidad'] == 'gramos') echo 'selected'; ?>>Gramos (g)</option>
                <option value="kilogramos" <?php if ($material['unidad'] == 'kilogramos') echo 'selected'; ?>>Kilogramos (kg)</option>
                <option value="litros" <?php if ($material['unidad'] == 'litros') echo 'selected'; ?>>Litros (l)</option>
                <option value="mililitros" <?php if ($material['unidad'] == 'mililitros') echo 'selected'; ?>>Mililitros (ml)</option>
                <option value="milimetros" <?php if ($material['unidad'] == 'milimetros') echo 'selected'; ?>>Milímetros (mm)</option>
                <option value="centimetros" <?php if ($material['unidad'] == 'centimetros') echo 'selected'; ?>>Centímetros (cm)</option>
                <option value="metros" <?php if ($material['unidad'] == 'metros') echo 'selected'; ?>>Metros (m)</option>
            </optgroup>
        </select><br><br>

        <label for="cantidad_disponible">Cantidad disponible:</label><br>
        <input type="number" id="cantidad_disponible" name="cantidad_disponible" min="0" step="1" value="<?php echo htmlspecialchars($material['cantidad_disponible']); ?>"><br><br>

        <label for="cantidad_minima">Cantidad mínima:</label><br>
        <input type="number" id="cantidad_minima" name="cantidad_minima" min="0" step="1" value="<?php echo htmlspecialchars($material['cantidad_minima']); ?>"><br><br>

        <label for="notas">Notas adicionales:</label><br>
        <textarea id="notas" name="notas" rows="4" cols="50"><?php echo htmlspecialchars($material['notas']); ?></textarea><br><br>

        <br>
        <button type="submit">Guardar Cambios</button>
    </fieldset>
</form>

<?php
require_once "../_incl/post-body.php";
?>
