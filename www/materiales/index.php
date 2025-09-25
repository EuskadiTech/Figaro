<?php
require_once "../_incl/utils.php";
require_permission("materiales.index");
require_once "../_incl/pre-body.php";

$centro = get_selected_centro();
$materiales_path = "$RUTA_DATOS/Centros/$centro/Materiales";
$materiales = [];

if (file_exists($materiales_path)) {
    $files = glob($materiales_path . '/*.json');
    foreach ($files as $file) {
        $data = json_decode(file_get_contents($file), true);
        if ($data) {
            $data['id'] = basename($file);
            $materiales[] = $data;
        }
    }
}
?>

<h1>Inventario de Materiales</h1>

<a href="crear_material.php" class="button"><img loading="lazy" class="picto" src="/static/pictos/nuevo.png"><br>Añadir Material</a>

<div class="materiales-list">
    <?php if (empty($materiales)): ?>
        <p>No hay materiales en el inventario.</p>
    <?php else: ?>
        <table>
            <thead>
                <tr>
                    <th>Foto</th>
                    <th>Nombre</th>
                    <th>Cantidad Disponible</th>
                    <th>Cantidad Mínima</th>
                    <th>Acciones</th>
                </tr>
            </thead>
            <tbody>
                <?php foreach ($materiales as $material): ?>
                    <tr>
                        <td><img loading="lazy" src="<?php echo htmlspecialchars($material['foto']); ?>" alt="Foto" style="max-height: 50px;"></td>
                        <td><?php echo htmlspecialchars($material['nombre']); ?></td>
                        <td><?php echo htmlspecialchars($material['cantidad_disponible']); ?> <?php echo htmlspecialchars($material['unidad']); ?>s</td>
                        <td><?php echo htmlspecialchars($material['cantidad_minima']); ?></td>
                        <td>
                            <a class="button" href="editar_material.php?id=<?php echo urlencode($material['id']); ?>"><img loading="lazy" class="picto" src="/static/pictos/escribir.png"><br>Editar</a>
                            <a class="button rojo" href="eliminar_material.php?id=<?php echo urlencode($material['id']); ?>"><img loading="lazy" class="picto" src="/static/pictos/borrar.png"><br>Eliminar</a>
                        </td>
                    </tr>
                <?php endforeach; ?>
            </tbody>
        </table>
    <?php endif; ?>
</div>

<?php
require_once "../_incl/post-body.php";
?>
