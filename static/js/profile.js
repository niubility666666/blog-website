

// 账户信息保存功能
document.getElementById('profileForm').addEventListener('submit', function(e) {
    e.preventDefault();

    const formData = new FormData(this);
    const userData = {
        motto: formData.get('motto'),
        github: formData.get('github'),
        google_account: formData.get('googleAccount')
    };

    fetch('/api/users/profile', {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(userData)
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                customAlert.success('个人信息更新成功');
            } else {
                customAlert.error('更新失败: ' + data.message);
            }
        })
        .catch(error => {
            console.error('Error:', error);
            customAlert.error('网络错误，请稍后重试');
        });
});

// 修改密码功能
document.getElementById('securityForm').addEventListener('submit', function(e) {
    e.preventDefault();

    const currentPassword = document.getElementById('currentPassword').value;
    const newPassword = document.getElementById('newPassword').value;
    const confirmPassword = document.getElementById('confirmPassword').value;

    // 密码确认验证
    if (newPassword !== confirmPassword) {
        customAlert.error('新密码与确认密码不一致');
        return;
    }

    // 密码强度验证（可选）
    if (newPassword.length < 6) {
        customAlert.error('密码长度至少6位');
        return;
    }

    const passwordData = {
        current_password: currentPassword,
        new_password: newPassword
    };

    fetch('/api/users/password', {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(passwordData)
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                customAlert.success('密码修改成功');
                // 清空密码输入框
                document.getElementById('securityForm').reset();
            } else {
                customAlert.error('密码修改失败: ' + data.message);
            }
        })
        .catch(error => {
            console.error('Error:', error);
            customAlert.error('网络错误，请稍后重试');
        });
});

// 切换密码可见性
function togglePasswordVisibility(inputId) {
    const input = document.getElementById(inputId);
    const toggleIcon = input.nextElementSibling;

    if (input.type === 'password') {
        input.type = 'text';
        // 可以在这里改变图标样式，表示当前是可见状态
        toggleIcon.innerHTML = `
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 2c2.5 0 4.5 1.8 5.5 4 .5 1 .5 2 0 3C12.5 11.2 10.5 13 8 13s-4.5-1.8-5.5-4c-.5-1-.5-2 0-3C3.5 3.8 5.5 2 8 2zm0 1.5c-2 0-3.7 1.5-4.5 3.5.8 2 2.5 3.5 4.5 3.5s3.7-1.5 4.5-3.5c-.8-2-2.5-3.5-4.5-3.5zm0 2a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3z"/>
                <path d="M10.5 8l2.5 2.5" stroke="currentColor" stroke-width="1" fill="none"/>
                <path d="M5.5 8l-2.5 2.5" stroke="currentColor" stroke-width="1" fill="none"/>
            </svg>
        `;
    } else {
        input.type = 'password';
        // 恢复原始图标样式
        toggleIcon.innerHTML = `
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 2c2.5 0 4.5 1.8 5.5 4 .5 1 .5 2 0 3C12.5 11.2 10.5 13 8 13s-4.5-1.8-5.5-4c-.5-1-.5-2 0-3C3.5 3.8 5.5 2 8 2zm0 1.5c-2 0-3.7 1.5-4.5 3.5.8 2 2.5 3.5 4.5 3.5s3.7-1.5 4.5-3.5c-.8-2-2.5-3.5-4.5-3.5zm0 2a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3z"/>
            </svg>
        `;
    }
}