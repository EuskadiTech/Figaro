<?php
require_once "../_incl/utils.php";
require_permission("ADMIN");
require_once "../_incl/pre-body.php";

$centros = get_centros();
$report_data = [];

foreach ($centros as $centro) {
    $materiales_path = "$RUTA_DATOS/Centros/$centro/Materiales";
    if (file_exists($materiales_path)) {
        $files = glob($materiales_path . '/*.json');
        foreach ($files as $file) {
            $data = json_decode(file_get_contents($file), true);
            if ($data) {
                $data['id'] = basename($file);
                $report_data[$centro][] = $data;
            }
        }
    }
}
?>

<h1>Informe de Materiales por Centro</h1>
<button onclick="window.print()" class="no_print">Imprimir Informe</button>

<br>
<span style="background-color: #ccccff;">Materiales con bajo stock</span>
<span style="background-color: #ffcccc;">Materiales sin stock</span>

<br><br>
<div class="report">
    <?php if (empty($report_data)): ?>
        <p>No hay materiales en el inventario.</p>
    <?php else: ?>
        <table>
            <thead>
                <tr>
                    <th>Foto</th>
                    <th>Nombre</th>
                    <th>Cantidad Disponible</th>
                    <th>Cantidad Mínima</th>
                </tr>
            </thead>
            <tbody>
                <?php foreach ($report_data as $centro => $materiales): ?>
                    <tr><td colspan="5" style="background-color: #f0f0f0; font-weight: bold; text-align: center;"><?php echo htmlspecialchars($centro); ?></td></tr>
                    <?php foreach ($materiales as $material): ?>
                        <tr style="<?php if ($material['cantidad_disponible'] == 0) echo 'background-color: #ffcccc;'; elseif ($material['cantidad_disponible'] < $material['cantidad_minima']) echo 'background-color: #ccccff;'; ?>">
                            <td><img loading="lazy" src="<?php echo htmlspecialchars($material['foto']); ?>" alt="Foto" style="max-height: 50px;"></td>
                            <td><?php echo htmlspecialchars($material['nombre']); ?></td>
                            <td><?php echo htmlspecialchars($material['cantidad_disponible']); ?> <?php echo htmlspecialchars($material['unidad']); ?>s</td>
                            <td><?php echo htmlspecialchars($material['cantidad_minima']); ?></td>
                        </tr>
                    <?php endforeach; ?>
                <?php endforeach; ?>
            </tbody>
        </table>
    <?php endif; ?>
</div>

<?php
require_once "../_incl/post-body.php";
?>
