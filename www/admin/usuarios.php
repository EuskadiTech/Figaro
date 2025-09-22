<?php
$SKIP_CENTRO = true; // Admin doesn't need centro/aula selection
require_once "../_incl/utils.php";
require_permission("ADMIN");
require_once "../_incl/pre-body.php";

$message = '';
$error = '';

// Handle form submissions
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    if (isset($_POST['action'])) {
        switch ($_POST['action']) {
            case 'create':
                $username = trim($_POST['username'] ?? '');
                $display_name = trim($_POST['display_name'] ?? '');
                $email = trim($_POST['email'] ?? '');
                $password = $_POST['password'] ?? '';
                $auth = $_POST['auth'] ?? [];
                
                if (empty($username) || empty($password)) {
                    $error = "Username and password are required.";
                } elseif (load_user($username)) {
                    $error = "User already exists.";
                } else {
                    $user_data = [
                        'password' => password_hash($password, PASSWORD_DEFAULT),
                        'auth' => $auth,
                        'display_name' => $display_name,
                        'email' => $email,
                        'created_at' => date('c')
                    ];
                    if (save_user($username, $user_data)) {
                        $message = "User created successfully.";
                    } else {
                        $error = "Failed to create user.";
                    }
                }
                break;
                
            case 'update':
                $username = $_POST['username'] ?? '';
                $user = load_user($username);
                if (!$user) {
                    $error = "User not found.";
                } else {
                    $user['display_name'] = trim($_POST['display_name'] ?? $user['display_name']);
                    $user['email'] = trim($_POST['email'] ?? $user['email']);
                    $user['auth'] = $_POST['auth'] ?? $user['auth'];
                    
                    // Only update password if provided
                    if (!empty($_POST['password'])) {
                        $user['password'] = password_hash($_POST['password'], PASSWORD_DEFAULT);
                    }
                    
                    $user['updated_at'] = date('c');
                    
                    if (save_user($username, $user)) {
                        $message = "User updated successfully.";
                    } else {
                        $error = "Failed to update user.";
                    }
                }
                break;
                
            case 'delete':
                $username = $_POST['username'] ?? '';
                if ($username === $_COOKIE['username']) {
                    $error = "Cannot delete your own account.";
                } elseif (delete_user($username)) {
                    $message = "User deleted successfully.";
                } else {
                    $error = "Failed to delete user.";
                }
                break;
        }
    }
}

$users = get_all_users();
$available_permissions = [
    'ADMIN' => 'Administrator (full access)',
    'materiales.index' => 'View materials',
    'materiales.create' => 'Create materials',
    'materiales.update' => 'Edit materials',
    'materiales.delete' => 'Delete materials',
    'actividades.index' => 'View activities',
    'archivos.index' => 'View files',
    'module2' => 'Module 2 access'
];
?>

<h1>Gestión de Usuarios</h1>

<?php if ($message): ?>
    <div style="background: lightgreen; border: 1px solid green; padding: 10px; margin: 10px 0; border-radius: 5px;">
        <?php echo htmlspecialchars($message); ?>
    </div>
<?php endif; ?>

<?php if ($error): ?>
    <div style="background: lightcoral; border: 1px solid red; padding: 10px; margin: 10px 0; border-radius: 5px;">
        <?php echo htmlspecialchars($error); ?>
    </div>
<?php endif; ?>

<div style="margin-bottom: 20px;">
    <button onclick="toggleCreateForm()" class="button">Crear Nuevo Usuario</button>
</div>

<!-- Create User Form -->
<div id="createUserForm" style="display: none; background: #f0f0f0; padding: 15px; border-radius: 5px; margin-bottom: 20px;">
    <h3>Crear Nuevo Usuario</h3>
    <form method="POST">
        <input type="hidden" name="action" value="create">
        
        <div style="margin-bottom: 10px;">
            <label for="new_username">Nombre de usuario:</label><br>
            <input type="text" id="new_username" name="username" required style="width: 200px;">
        </div>
        
        <div style="margin-bottom: 10px;">
            <label for="new_display_name">Nombre completo:</label><br>
            <input type="text" id="new_display_name" name="display_name" style="width: 200px;">
        </div>
        
        <div style="margin-bottom: 10px;">
            <label for="new_email">Email:</label><br>
            <input type="email" id="new_email" name="email" style="width: 200px;">
        </div>
        
        <div style="margin-bottom: 10px;">
            <label for="new_password">Contraseña:</label><br>
            <input type="password" id="new_password" name="password" required style="width: 200px;">
        </div>
        
        <div style="margin-bottom: 10px;">
            <label>Permisos:</label><br>
            <?php foreach ($available_permissions as $perm => $desc): ?>
                <label style="display: block; margin: 5px 0;">
                    <input type="checkbox" name="auth[]" value="<?php echo htmlspecialchars($perm); ?>">
                    <?php echo htmlspecialchars($desc); ?>
                </label>
            <?php endforeach; ?>
        </div>
        
        <button type="submit" class="button">Crear Usuario</button>
        <button type="button" onclick="toggleCreateForm()" class="button">Cancelar</button>
    </form>
</div>

<!-- Users List -->
<h3>Usuarios Existentes</h3>
<table style="width: 100%; border-collapse: collapse;">
    <thead>
        <tr style="background: #f0f0f0;">
            <th style="border: 1px solid #ccc; padding: 10px; text-align: left;">Usuario</th>
            <th style="border: 1px solid #ccc; padding: 10px; text-align: left;">Nombre</th>
            <th style="border: 1px solid #ccc; padding: 10px; text-align: left;">Email</th>
            <th style="border: 1px solid #ccc; padding: 10px; text-align: left;">Permisos</th>
            <th style="border: 1px solid #ccc; padding: 10px; text-align: left;">Acciones</th>
        </tr>
    </thead>
    <tbody>
        <?php foreach ($users as $username => $user): ?>
            <tr>
                <td style="border: 1px solid #ccc; padding: 10px;"><?php echo htmlspecialchars($username); ?></td>
                <td style="border: 1px solid #ccc; padding: 10px;"><?php echo htmlspecialchars($user['display_name'] ?? ''); ?></td>
                <td style="border: 1px solid #ccc; padding: 10px;"><?php echo htmlspecialchars($user['email'] ?? ''); ?></td>
                <td style="border: 1px solid #ccc; padding: 10px;"><?php echo htmlspecialchars(implode(', ', $user['auth'] ?? [])); ?></td>
                <td style="border: 1px solid #ccc; padding: 10px;">
                    <button onclick="editUser('<?php echo htmlspecialchars($username); ?>')" class="button">Editar</button>
                    <?php if ($username !== $_COOKIE['username']): ?>
                        <button onclick="deleteUser('<?php echo htmlspecialchars($username); ?>')" class="button rojo">Eliminar</button>
                    <?php endif; ?>
                </td>
            </tr>
            
            <!-- Edit Form (hidden by default) -->
            <tr id="editForm_<?php echo htmlspecialchars($username); ?>" style="display: none;">
                <td colspan="5" style="border: 1px solid #ccc; padding: 15px; background: #f9f9f9;">
                    <form method="POST">
                        <input type="hidden" name="action" value="update">
                        <input type="hidden" name="username" value="<?php echo htmlspecialchars($username); ?>">
                        
                        <div style="display: inline-block; margin-right: 20px;">
                            <label>Nombre completo:</label><br>
                            <input type="text" name="display_name" value="<?php echo htmlspecialchars($user['display_name'] ?? ''); ?>" style="width: 200px;">
                        </div>
                        
                        <div style="display: inline-block; margin-right: 20px;">
                            <label>Email:</label><br>
                            <input type="email" name="email" value="<?php echo htmlspecialchars($user['email'] ?? ''); ?>" style="width: 200px;">
                        </div>
                        
                        <div style="display: inline-block; margin-right: 20px;">
                            <label>Nueva contraseña (dejar vacío para no cambiar):</label><br>
                            <input type="password" name="password" style="width: 200px;">
                        </div>
                        
                        <div style="margin-top: 10px;">
                            <label>Permisos:</label><br>
                            <?php foreach ($available_permissions as $perm => $desc): ?>
                                <label style="display: block; margin: 5px 0;">
                                    <input type="checkbox" name="auth[]" value="<?php echo htmlspecialchars($perm); ?>" 
                                           <?php echo in_array($perm, $user['auth'] ?? []) ? 'checked' : ''; ?>>
                                    <?php echo htmlspecialchars($desc); ?>
                                </label>
                            <?php endforeach; ?>
                        </div>
                        
                        <div style="margin-top: 15px;">
                            <button type="submit" class="button">Guardar Cambios</button>
                            <button type="button" onclick="cancelEdit('<?php echo htmlspecialchars($username); ?>')" class="button">Cancelar</button>
                        </div>
                    </form>
                </td>
            </tr>
        <?php endforeach; ?>
    </tbody>
</table>

<script>
function toggleCreateForm() {
    const form = document.getElementById('createUserForm');
    form.style.display = form.style.display === 'none' ? 'block' : 'none';
}

function editUser(username) {
    const editForm = document.getElementById('editForm_' + username);
    editForm.style.display = editForm.style.display === 'none' ? 'table-row' : 'none';
}

function cancelEdit(username) {
    const editForm = document.getElementById('editForm_' + username);
    editForm.style.display = 'none';
}

function deleteUser(username) {
    if (confirm('¿Estás seguro de que quieres eliminar el usuario "' + username + '"?')) {
        const form = document.createElement('form');
        form.method = 'POST';
        form.innerHTML = '<input type="hidden" name="action" value="delete"><input type="hidden" name="username" value="' + username + '">';
        document.body.appendChild(form);
        form.submit();
    }
}
</script>

<?php require_once "../_incl/post-body.php"; ?>