// 发表评论功能
document.addEventListener('DOMContentLoaded', function() {
    // 获取评论表单元素
    const commentForm = document.querySelector('.comment-form');
    const commentTextarea = commentForm.querySelector('textarea');
    const submitButton = commentForm.querySelector('.btn-primary');
    const parentIdInput = document.getElementById('parent-id');
    const postId = parseInt(commentForm.dataset.postId) || 0;
    // 获取parent_id值
    const parentId = parentIdInput ? parseInt(parentIdInput.value) || 0 : 0;

    // 监听发表评论按钮点击事件
    submitButton.addEventListener('click', function(e) {
        e.preventDefault(); // 阻止默认表单提交行为

        // 获取textarea内容
        const commentContent = commentTextarea.value.trim();

        // 验证内容是否为空
        if (!commentContent) {
            // alert('请输入评论内容');
            customAlert.error('请输入评论内容');
            return;
        }

        // 构造提交数据
        const postData = {
            content: commentContent,
            post_id: postId, // 从模板获取文章ID
            parent_id: parentId // 从模板获取文章ID
        };

        // 提交数据到服务器
        fetch('/api/comments', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify(postData)
        })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    // 清空textarea
                    commentTextarea.value = '';
                    if (parentIdInput) {
                        parentIdInput.value = '0';
                    }
                    // 显示成功消息
                    // alert('评论发表成功');
                    customAlert.success('评论发表成功', 3500);
                    // 这里可以考虑重新加载评论列表或动态添加评论
                    location.reload(); // 简单处理，刷新页面
                } else {
                    // alert('评论发表失败: ' + data.message);
                    customAlert.error('评论发表失败: ' + data.message);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                // alert('网络错误，请稍后再试');
                customAlert.error('网络错误，请稍后重试');
            });
    });

    // 添加回复按钮事件监听器
    document.querySelectorAll('.reply-btn').forEach(button => {
        button.addEventListener('click', function() {
            // 获取被回复的评论ID
            const commentItem = this.closest('.comment-item');
            const commentId = commentItem.dataset.commentId ||
                commentItem.parentElement.closest('.comment-item')?.dataset.commentId ||
                commentItem.id?.replace('comment-', '') ||
                0;

            const commentAuthorId = commentItem.dataset.userId;
            const currentUserId = commentItem.dataset.currentUserId;

            if (currentUserId && commentAuthorId && currentUserId === commentAuthorId) {
                customAlert.error('不能回复自己的评论');
                return;
            }

            // 设置parent_id
            if (parentIdInput) {
                parentIdInput.value = commentId;
            }

            // 聚焦到评论框
            commentTextarea.focus();

            // 可选：在评论框中添加@用户名提示
            const authorName = commentItem.querySelector('.comment-author')?.textContent || '';
            if (authorName && commentTextarea.value.indexOf(`@${authorName}`) === -1) {
                commentTextarea.value = `@${authorName} ` + `#${commentId} ` + commentTextarea.value;
            }
        });
    });
});

// 添加文章点赞功能
document.addEventListener('DOMContentLoaded', function() {
    // 文章点赞按钮事件监听
    const postLikeBtn = document.querySelector('.post-actions .like-btn');
    if (postLikeBtn) {
        postLikeBtn.addEventListener('click', function(e) {
            e.preventDefault();

            // 获取文章ID
            const postId = document.querySelector('.comment-form')?.dataset.postId;
            if (!postId) {
                customAlert.error('无法获取文章信息');
                return;
            }

            // 判断当前是点赞还是取消点赞
            const isLiked = this.classList.contains('liked');
            const action = isLiked ? 'unlike' : 'like';

            // 发送请求
            fetch(`/api/posts/${postId}/like`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
                },
                body: JSON.stringify({
                    action: action
                })
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // 更新界面
                        const likeCountElement = this.querySelector('span:last-child');
                        if (likeCountElement) {
                            likeCountElement.textContent = data.likes;
                        }

                        // 切换按钮状态
                        if (action === 'like') {
                            this.classList.add('liked');
                            customAlert.success('点赞成功');
                        } else {
                            this.classList.remove('liked');
                            customAlert.error('取消点赞');
                        }


                    } else {
                        customAlert.error(data.message || '操作失败');
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    customAlert.error('网络错误，请稍后重试');
                });
        });
    }
});


// 添加点赞按钮事件监听器
document.addEventListener('DOMContentLoaded', function() {
    // 为所有点赞按钮添加事件监听
    document.querySelectorAll('.comment-actions .like-btn').forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();

            const commentId = this.dataset.commentId;
            const action = this.dataset.action;

            // 发送点赞请求到后端
            fetch(`/api/comments/${commentId}/like`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
                },
                body: JSON.stringify({
                    action: action
                })
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // 更新点赞数显示
                        const likeCount = parseInt(this.textContent.match(/\d+/)?.[0] || '0');
                        if (action === 'like') {
                            this.innerHTML = `👍 ${likeCount + 1}`;
                            // 防止重复点赞，可以禁用按钮或改变样式
                            this.dataset.action = 'unlike';
                        } else {
                            this.innerHTML = `👍 ${likeCount - 1}`;
                            this.dataset.action = 'like';
                        }
                    } else {
                        // alert('操作失败: ' + data.message);
                        customAlert.error('操作失败: ' + data.message);
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    // alert('网络错误，请稍后再试');
                    customAlert.error('网络错误，请稍后重试');
                });
        });
    });
});

// 分享功能实现 - 修复版本
document.addEventListener('DOMContentLoaded', function() {
    // 获取分享按钮
    const shareBtn = document.querySelector('.post-actions .share-btn');
    // 添加分享状态标识
    let isSharing = false;

    if (shareBtn) {
        shareBtn.addEventListener('click', function() {
            // 检查是否正在进行分享
            if (isSharing) {
                customAlert.error('分享正在进行中，请稍候...');
                return;
            }

            // 获取文章标题和URL
            const title = document.querySelector('.post-title')?.textContent || '';
            const url = window.location.href;

            // 构造分享文本
            const shareText = `推荐文章：${title}`;
            const isMobile = /Android|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
            // 尝试使用Web Share API（移动端）
            if (isMobile) {
                isSharing = true;
                navigator.share({
                    title: title,
                    text: shareText,
                    url: url
                })
                    .then(() => {
                        customAlert.success('分享成功');
                    })
                    .catch((error) => {
                        if (error.name !== 'AbortError') {
                            console.error('分享失败:', error);
                            showShareOptions(title, url, shareText);
                        }
                    })
                    .finally(() => {
                        // 无论成功或失败都重置分享状态
                        isSharing = false;
                    });
            } else {
                // Web Share API不可用时，直接显示传统分享选项
                showShareOptions(title, url, shareText);
            }
        });
    }
});

// 显示分享选项对话框
function showShareOptions(title, url, text) {
    // 创建分享选项对话框
    const modal = document.createElement('div');
    modal.className = 'share-modal';
    modal.innerHTML = `
        <div class="share-overlay"></div>
        <div class="share-dialog">
            <div class="share-header">
                <h3>分享到</h3>
                <button class="share-close">&times;</button>
            </div>
            <div class="share-options">
                <button class="share-option" data-platform="copy">
                    <span class="icon">📋</span>
                    <span>复制链接</span>
                </button>
                <button class="share-option" data-platform="wechat">
                    <span class="icon">💬</span>
                    <span>微信</span>
                </button>
                <button class="share-option" data-platform="weibo">
                    <span class="icon">📊</span>
                    <span>微博</span>
                </button>
                <button class="share-option" data-platform="qq">
                    <span class="icon">🐧</span>
                    <span>QQ</span>
                </button>
            </div>
        </div>
    `;

    // 添加到页面
    document.body.appendChild(modal);

    // 关闭模态框
    const close = () => {
        document.body.removeChild(modal);
    };

    // 绑定关闭事件
    modal.querySelector('.share-overlay').addEventListener('click', close);
    modal.querySelector('.share-close').addEventListener('click', close);

    // 绑定分享选项事件
    modal.querySelectorAll('.share-option').forEach(option => {
        option.addEventListener('click', function() {
            const platform = this.dataset.platform;
            handleShare(platform, title, url, text);
            close();
        });
    });
}

// 处理不同平台的分享
function handleShare(platform, title, url, text) {
    switch(platform) {
        case 'copy':
            copyToClipboard(url);
            break;
        case 'wechat':
            customAlert.info('请在微信中打开链接进行分享');
            break;
        case 'weibo':
            window.open(`https://service.weibo.com/share/share.php?url=${encodeURIComponent(url)}&title=${encodeURIComponent(text)}`, '_blank');
            break;
        case 'qq':
            window.open(`https://connect.qq.com/widget/shareqq/index.html?url=${encodeURIComponent(url)}&title=${encodeURIComponent(title)}&desc=${encodeURIComponent(text)}`, '_blank');
            break;
    }
}

// 复制到剪贴板 - 修复版本
function copyToClipboard(text) {
    if (navigator.clipboard) {
        navigator.clipboard.writeText(text)
            .then(() => {
                customAlert.success('链接已复制到剪贴板');
            })
            .catch(err => {
                console.error('复制失败:', err);
                customAlert.error('复制失败');
            });
    } else {
        // 对于不支持 Clipboard API 的浏览器，使用现代替代方案
        fallbackCopyTextToClipboard(text);
    }
}

// 降级复制方法 - 使用现代方法
function fallbackCopyTextToClipboard(text) {
    try {
        // 创建临时输入元素
        const input = document.createElement('input');
        input.style.position = 'fixed';
        input.style.opacity = '0';
        input.value = text;
        document.body.appendChild(input);
        input.select();

        // 尝试执行复制命令
        const successful = document.execCommand('copy');
        document.body.removeChild(input);

        if (successful) {
            customAlert.success('链接已复制到剪贴板');
        } else {
            customAlert.error('复制失败');
        }
    } catch (err) {
        // 如果 execCommand 也失败，则显示错误
        console.error('复制失败:', err);
        customAlert.error('复制失败');

        // 最后的备选方案：提示用户手动复制
        prompt('请手动复制以下链接:', text);
    }
}


// 收藏功能实现
document.addEventListener('DOMContentLoaded', function() {
    // 获取收藏按钮
    const favoriteBtn = document.querySelector('.post-actions .favorite-btn');

    if (favoriteBtn) {
        favoriteBtn.addEventListener('click', function(e) {
            e.preventDefault();

            // 获取文章ID
            const postId = document.querySelector('.comment-form')?.dataset.postId;
            if (!postId) {
                customAlert.error('无法获取文章信息');
                return;
            }

            // 判断当前是收藏还是取消收藏
            const isFavorited = this.classList.contains('favorited');
            const action = isFavorited ? 'unfavorite' : 'favorite';

            // 发送请求到后端API（需要后端实现对应的API）
            fetch(`/api/posts/${postId}/favorite`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
                },
                body: JSON.stringify({
                    action: action
                })
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // 更新界面
                        if (action === 'favorite') {
                            this.classList.add('favorited');
                            this.innerHTML = '<span class="icon">⭐</span><span>已收藏</span>';
                            customAlert.success('收藏成功');
                        } else {
                            this.classList.remove('favorited');
                            this.innerHTML = '<span class="icon">⭐</span><span>收藏</span>';
                            customAlert.success('取消收藏成功');
                        }
                    } else {
                        customAlert.error(data.message || '操作失败');
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    customAlert.error('网络错误，请稍后重试');
                });
        });
    }
});


