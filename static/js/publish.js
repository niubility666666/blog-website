document.addEventListener('DOMContentLoaded', function (index) {
    // 初始化 CodeMirror 编辑器
    const editor = CodeMirror.fromTextArea(document.getElementById('editorContent'), {
        mode: 'markdown',
        theme: 'default',
        lineNumbers: true,
        lineWrapping: false,
        autofocus: false,
        viewportMargin: Infinity,
        extraKeys: {
            "Enter": "newlineAndIndentContinueMarkdownList"
        },
        placeholder: "鼓励友善发言，禁止人身攻击"
    });

    const markedParse = marked
    markedParse.setOptions({
        gfm: true,
        tables: true,
        escaped: true,
        breaks: false,
        pedantic: false,
        sanitize: false,
        smartLists: true,
        smartypants: false,
    })
    console.log(markedParse, 'jjj---')

    // 设置编辑器高度
    editor.setSize('100%', '300px');

    // 获取预览相关元素
    // 内容、预览、对照标签页切换功能
    const tabOptions = document.querySelectorAll('.tab-option');
    const editorPane = document.getElementById('editorPane');
    const editorWrapper = document.querySelector('.editor-wrapper');


    // 后续再添加事件监听器
    // setTimeout(() => {
    //     editor.on('change', function(cm) {
    //         console.log('编辑器内容已更改');
    //         htmlContent = marked.parse(cm.getValue())
    //         console.log(htmlContent, 'oooo----')
    //     });
    // }, 0);


    // 表单提交
    const publishForm = document.getElementById('publishForm');
    publishForm.addEventListener('submit', function (e) {
        e.preventDefault();

        // 获取表单数据
        const title = document.getElementById('title').value;
        const tags = document.getElementById('tags').value; // 隐藏字段，包含所有标签
        const content = editor.getValue(); // CodeMirror编辑器内容
        const htmlContent = marked.parse(content)
        const category = document.querySelector('select[name="category"]').value;
        const readLimit = document.querySelector('select[name="readLimit"]').value;

        const data = {
            title: title,
            tags: tags,
            content: htmlContent,
            category_id: parseInt(category),
            read_limit: parseInt(readLimit)
        };

        // 发送到后端
        fetch('/api/posts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data)
        })
            .then(response => response.json())
            .then(result => {
                if (result.success) {
                    customAlert.success('文章发布成功！');
                    window.location.href = '/';
                } else {
                    customAlert.error('发布失败: ' + result.message);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                customAlert.error('发布过程中出现错误');
            });
    });

    // 1、监听点击行号，点击事件设置编辑器行号显示与隐藏
    document.querySelector('.tab-option-has-icon[title="行号"]').addEventListener('click', function () {
        const isLineNumbersVisible = editor.getOption('lineNumbers');
        editor.setOption('lineNumbers', !isLineNumbersVisible);
    });

    // 2、监听点击工具栏，点击事件设置工具栏区域显示与隐藏
    document.querySelector('.tab-option-has-icon[title="工具栏"]').addEventListener('click', function () {
        const toolbar = document.querySelector('.mde-toolbar');
        const isVisible = toolbar.style.display !== 'none';
        toolbar.style.display = isVisible ? 'none' : 'flex';
    });

    // 3、监听点击全屏，点击事件设置编辑器全屏功能打开与关闭
    document.querySelector('.tab-option-has-icon[title="全屏"]').addEventListener('click', function () {
        const editorContainer = document.querySelector('.editor-wrapper');
        const isFullscreen = editorContainer.classList.contains('fullscreen');

        if (isFullscreen) {
            // 退出全屏模式
            editorContainer.classList.remove('fullscreen');
            editor.setSize('100%', '300px');
            // 恢复其他可能被隐藏的元素
            document.querySelector('.publish-header').style.display = '';
            document.querySelector('.publish-form-group[label="文章标题"]').style.display = '';
            document.querySelector('.form-row').style.display = '';
            document.querySelector('.publish-actions').style.display = '';
        } else {
            // 进入全屏模式
            editorContainer.classList.add('fullscreen');
            editor.setSize('100%', 'calc(100vh - 40px)');
            // 隐藏其他元素以获得更多空间
            document.querySelector('.publish-header').style.display = 'none';
            document.querySelectorAll('.publish-form-group:not(:has(.editor-wrapper))').forEach(el => {
                el.style.display = 'none';
            });
            document.querySelector('.form-row').style.display = 'none';
            document.querySelector('.publish-actions').style.display = 'none';
        }
    });

    // 4、监听点击加粗，点击事件设置编辑器内容加粗
    // 5、监听点击斜体，点击事件设置编辑器内容斜体
    // 6、监听点击删除线，点击事件设置编辑器内容删除线
    // 7、监听点击标题，点击事件设置编辑器内容标题
    // 8、监听点击有序列表，点击事件设置编辑器内容有序列表
    // 9、监听点击无序列表，点击事件设置编辑器内容无序列表
    // 10、监听点击引用，点击事件设置编辑器内容引用

    // 统一处理工具栏按钮点击事件
    function handleToolbarAction(action) {
        const cursor = editor.getCursor();
        const selection = editor.getSelection();

        switch (action) {
            case 'bold':
                if (selection) {
                    editor.replaceSelection(`**${selection}**`);
                } else {
                    editor.replaceSelection('****');
                    editor.setCursor({line: editor.getCursor().line, ch: editor.getCursor().ch - 2});
                }
                break;

            case 'italic':
                if (selection) {
                    editor.replaceSelection(`*${selection}*`);
                } else {
                    editor.replaceSelection('**');
                    editor.setCursor({line: editor.getCursor().line, ch: editor.getCursor().ch - 1});
                }
                break;

            case 'strikethrough':
                if (selection) {
                    editor.replaceSelection(`~~${selection}~~`);
                } else {
                    editor.replaceSelection('~~~~');
                    editor.setCursor({line: editor.getCursor().line, ch: editor.getCursor().ch - 2});
                }
                break;

            case 'heading':
                const lineContent = editor.getLine(cursor.line);
                if (lineContent.startsWith('#')) {
                    const newContent = lineContent.replace(/^(#+\s*)/, '');
                    editor.replaceRange(newContent, {line: cursor.line, ch: 0}, {
                        line: cursor.line,
                        ch: lineContent.length
                    });
                } else {
                    editor.replaceRange(`# ${lineContent}`, {line: cursor.line, ch: 0}, {
                        line: cursor.line,
                        ch: lineContent.length
                    });
                }
                break;

            case 'orderedList':
                const olLineContent = editor.getLine(cursor.line);
                if (/^\d+\.\s/.test(olLineContent)) {
                    const newContent = olLineContent.replace(/^\d+\.\s/, '');
                    editor.replaceRange(newContent, {line: cursor.line, ch: 0}, {
                        line: cursor.line,
                        ch: olLineContent.length
                    });
                } else {
                    editor.replaceRange(`1. ${olLineContent}`, {line: cursor.line, ch: 0}, {
                        line: cursor.line,
                        ch: olLineContent.length
                    });
                }
                break;

            case 'unorderedList':
                const ulLineContent = editor.getLine(cursor.line);
                if (/^[\*\-\+]\s/.test(ulLineContent)) {
                    const newContent = ulLineContent.replace(/^[\*\-\+]\s/, '');
                    editor.replaceRange(newContent, {line: cursor.line, ch: 0}, {
                        line: cursor.line,
                        ch: ulLineContent.length
                    });
                } else {
                    editor.replaceRange(`- ${ulLineContent}`, {line: cursor.line, ch: 0}, {
                        line: cursor.line,
                        ch: ulLineContent.length
                    });
                }
                break;

            case 'quote':
                const qLineContent = editor.getLine(cursor.line);
                if (qLineContent.startsWith('> ')) {
                    const newContent = qLineContent.replace(/^>\s/, '');
                    editor.replaceRange(newContent, {line: cursor.line, ch: 0}, {
                        line: cursor.line,
                        ch: qLineContent.length
                    });
                } else {
                    editor.replaceRange(`> ${qLineContent}`, {line: cursor.line, ch: 0}, {
                        line: cursor.line,
                        ch: qLineContent.length
                    });
                }
                break;

            case 'link':
                if (selection) {
                    editor.replaceSelection(`[${selection}](url)`);
                } else {
                    editor.replaceSelection('[文本](url)');
                    editor.setCursor({line: editor.getCursor().line, ch: editor.getCursor().ch - 4});
                }
                break;

            case 'image':
                editor.replaceSelection('![alt text](image-url)');
                break;

            case 'codeBlock':
                editor.replaceSelection('```\n代码内容\n```')
                break;

            case 'table':
                const table = '| 列1 | 列2 | 列3 |\n| --- | --- | --- |\n| 内容1 | 内容2 | 内容3 |';
                editor.replaceSelection(table);
                break;

            case 'horizontalRule':
                editor.replaceSelection('\n---\n');
                break;

            case 'undo':
                editor.undo();
                break;

            case 'redo':
                editor.redo();
                break;

            case 'clear':
                editor.setValue('');
                break;
        }

        editor.focus();
    }

    // 4、监听点击加粗，点击事件设置编辑器内容加粗
    document.querySelector('.toolbar-item[title="加粗"]').addEventListener('click', () => handleToolbarAction('bold'));
    // 5、监听点击斜体，点击事件设置编辑器内容斜体
    document.querySelector('.toolbar-item[title="斜体"]').addEventListener('click', () => handleToolbarAction('italic'));
    // 6、监听点击删除线，点击事件设置编辑器内容删除线
    document.querySelector('.toolbar-item[title="删除线"]').addEventListener('click', () => handleToolbarAction('strikethrough'));
    // 7、监听点击标题，点击事件设置编辑器内容标题
    document.querySelector('.toolbar-item[title="标题"]').addEventListener('click', () => handleToolbarAction('heading'));
    // 8、监听点击有序列表，点击事件设置编辑器内容有序列表
    document.querySelector('.toolbar-item[title="有序列表"]').addEventListener('click', () => handleToolbarAction('orderedList'));
    // 9、监听点击无序列表，点击事件设置编辑器内容无序列表
    document.querySelector('.toolbar-item[title="无序列表"]').addEventListener('click', () => handleToolbarAction('unorderedList'));
    // 10、监听点击引用，点击事件设置编辑器内容引用
    document.querySelector('.toolbar-item[title="引用"]').addEventListener('click', () => handleToolbarAction('quote'));
    // 11、监听点击链接，点击事件设置编辑器内容链接
    document.querySelector('.toolbar-item[title="链接"]').addEventListener('click', () => handleToolbarAction('link'));
    // 12、监听点击图片，点击事件设置编辑器内容图片
    document.querySelector('.toolbar-item[title="图片"]').addEventListener('click', () => handleToolbarAction('image'));
    // 13、监听点击代码块，点击事件设置编辑器内容代码块
    document.querySelector('.toolbar-item[title="代码块"]').addEventListener('click', () => handleToolbarAction('codeBlock'));
    // 14、监听点击表格，点击事件设置编辑器内容表格
    document.querySelector('.toolbar-item[title="表格"]').addEventListener('click', () => handleToolbarAction('table'));
    // 15、监听点击分割线，点击事件设置编辑器内容分割线
    document.querySelector('.toolbar-item[title="分割线"]').addEventListener('click', () => handleToolbarAction('horizontalRule'));
    // 16、监听点击撤销，点击事件设置编辑器内容撤销
    document.querySelector('.toolbar-item[title="撤销"]').addEventListener('click', () => handleToolbarAction('undo'));
    // 17、监听点击重做，点击事件设置编辑器内容重做
    document.querySelector('.toolbar-item[title="重做"]').addEventListener('click', () => handleToolbarAction('redo'));
    // 18、监听点击清空，点击事件设置编辑器内容清空
    document.querySelector('.toolbar-item[title="清空"]').addEventListener('click', () => handleToolbarAction('clear'));


// 创建预览区域
    let previewPane = document.getElementById('editorPreview');
    if (!previewPane) {
        previewPane = document.createElement('div');
        previewPane.id = 'editorPreview';
        previewPane.className = 'editor-preview';
        previewPane.style.display = 'none';
        previewPane.style.height = '300px'; // 设置固定高度
        previewPane.style.overflow = 'auto';
        editorPane.parentNode.appendChild(previewPane);
    }

    tabOptions.forEach((tab, index) => {
        if (index < 3) { // 只处理前三个标签
            tab.addEventListener('click', function() {
                // 移除所有标签的激活状态
                tabOptions.forEach(t => t.classList.remove('tab-option-on'));
                // 添加当前标签的激活状态
                this.classList.add('tab-option-on');

                // 根据点击的标签显示对应内容
                switch(index) {
                    case 0: // 内容
                        editorPane.style.display = 'block';
                        previewPane.style.display = 'none';
                        editorWrapper.style.height = '380px';
                        editorPane.style.width = '100%';
                        break;
                    case 1: // 预览
                        editorPane.style.display = 'none';
                        previewPane.style.display = 'block';
                        editorWrapper.style.height = '380px';
                        previewPane.style.width = '100%';
                        previewPane.style.padding = '10px';
                        updatePreview();
                        break;
                    case 2: // 对照
                        editorPane.style.display = 'block';
                        previewPane.style.display = 'block';
                        editorWrapper.style.height = '380px';
                        // 重置浮动样式
                        editorPane.style.cssFloat = 'left';
                        editorPane.style.width = '50%';
                        previewPane.style.cssFloat = 'right';
                        previewPane.style.width = '50%';
                        // 确保预览区域正确显示
                        previewPane.style.height = '300px';
                        previewPane.style.overflow = 'auto';
                        // 清除可能存在的其他样式
                        previewPane.style.position = 'relative';
                        previewPane.style.border = '1px solid var(--card-bg)'
                        updatePreview();
                        break;
                }
            });
        }
    });

    // 更新预览内容的函数
    function updatePreview() {
        const markdownContent = editor.getValue();
        const htmlContent = marked.parse(markdownContent);
        if (previewPane) {
            previewPane.innerHTML = htmlContent;
            // 为表格添加样式
            const tables = previewPane.querySelectorAll('table');
            tables.forEach(table => {
                table.style.borderCollapse = 'collapse';
                table.style.width = '60%';
                table.style.margin = '8px 0';
                table.style.fontSize = '12px'; // 更小的字体

                // 为表格添加边框样式
                const allTh = table.querySelectorAll('th');
                const allTd = table.querySelectorAll('td');

                allTh.forEach(th => {
                    th.style.border = '1px solid var(--table-border-color, #ddd)';
                    th.style.padding = '6px 4px'; // 更小的内边距
                    th.style.backgroundColor = 'var(--table-header-bg, #f2f2f2)';
                    th.style.fontWeight = 'bold';
                    th.style.textAlign = 'left';
                    th.style.color = 'var(--table-header-text, #333)';
                });

                allTd.forEach(td => {
                    td.style.border = '1px solid var(--table-border-color, #ddd)';
                    td.style.padding = '6px 4px'; // 更小的内边距
                    td.style.verticalAlign = 'top';
                    td.style.color = 'var(--text-color, #333)';
                });

                // 为表格行添加交替背景色
                const rows = table.querySelectorAll('tr');
                rows.forEach((row, index) => {
                    if (index % 2 === 0 && index > 0) {
                        row.style.backgroundColor = 'var(--table-alt-row-bg, #f9f9f9)';
                    }
                });
            });
        }
    }



});


document.addEventListener('DOMContentLoaded', function() {
    const tagInput = document.getElementById('tagInput');
    const addTagBtn = document.getElementById('addTagBtn');
    const tagList = document.getElementById('tagList');
    const tagsHiddenInput = document.getElementById('tags');

    let tags = [];

    tagInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            addTag(tagInput.value.trim());
        }
    });

    addTagBtn.addEventListener('click', function() {
        addTag(tagInput.value.trim());
    });

    function addTag(tagText) {
        if (tagText && !tags.includes(tagText)) {
            tags.push(tagText);

            const tagElement = document.createElement('span');
            tagElement.className = 'tag-item';
            tagElement.innerHTML = `
                ${tagText}
                <span class="tag-remove" data-tag="${tagText}">&times;</span>
            `;

            tagList.appendChild(tagElement);

            updateTagsInput();

            tagInput.value = '';

            tagElement.querySelector('.tag-remove').addEventListener('click', function() {
                removeTag(tagText);
            });
        }
    }

    function removeTag(tagText) {
        tags = tags.filter(tag => tag !== tagText);

        const tagElements = tagList.querySelectorAll('.tag-item');
        tagElements.forEach(element => {
            if (element.textContent.includes(tagText)) {
                element.remove();
            }
        });

        updateTagsInput();
    }

    function updateTagsInput() {
        tagsHiddenInput.value = JSON.stringify(tags);
    }
});

