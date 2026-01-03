// 自定义Alert组件
class CustomAlert {
  constructor() {
    this.container = document.createElement('div');
    this.container.id = 'custom-alert-container'
    this.container.className = 'custom-alert-container';
    document.body.appendChild(this.container);
  }

  show(message, type = 'info', duration = 3000) {
    // 创建alert元素
    const alert = document.createElement('div');
    alert.className = `custom-alert ${type}`;

    // 根据类型设置图标
    let icon = 'ℹ️';
    if (type === 'success') icon = '✅';
    else if (type === 'error') icon = '❌';
    else if (type === 'warning') icon = '⚠️';

    alert.innerHTML = `
      <div class="alert-content">
        <span class="alert-icon">${icon}</span>
        <span class="alert-message">${message}</span>
        <button class="alert-close">&times;</button>
      </div>
    `;

    // 添加关闭事件
    const closeBtn = alert.querySelector('.alert-close');
    closeBtn.addEventListener('click', () => {
      this.hide(alert);
    });

    // 添加到容器
    this.container.appendChild(alert);

    // 触发显示动画
    setTimeout(() => {
      alert.classList.add('show');
    }, 10);

    // 自动关闭
    if (duration > 0) {
      setTimeout(() => {
        this.hide(alert);
      }, duration);
    }

    return alert;
  }

  hide(alert) {
    alert.classList.remove('show');
    setTimeout(() => {
      if (alert.parentNode) {
        alert.parentNode.removeChild(alert);
      }
    }, 300);
  }

  success(message, duration) {
    return this.show(message, 'success', duration);
  }

  error(message, duration) {
    return this.show(message, 'error', duration);
  }

  warning(message, duration) {
    return this.show(message, 'warning', duration);
  }

  info(message, duration) {
    return this.show(message, 'info', duration);
  }
}

// 创建全局实例
const customAlert = new CustomAlert();

// 主题切换功能
class ThemeManager {
  constructor() {
    this.themeToggle = document.getElementById('themeToggle');
    this.themeIcon = this.themeToggle.querySelector('.theme-icon');
    this.body = document.body;

    this.init();
  }

  init() {
    // 从localStorage加载用户主题偏好
    const savedTheme = localStorage.getItem('theme') || 'dark-theme';
    this.setTheme(savedTheme);

    // 绑定切换事件
    this.themeToggle.addEventListener('click', () => this.toggleTheme());

    // 添加主题切换动画类
    this.body.classList.add('theme-transition');
  }

  toggleTheme() {
    const isDark = this.body.classList.contains('dark-theme');
    const newTheme = isDark ? 'light-theme' : 'dark-theme';
    this.setTheme(newTheme);
  }

  setTheme(theme) {
    // 移除现有主题类
    this.body.classList.remove('dark-theme', 'light-theme');

    // 添加新主题类
    this.body.classList.add(theme);

    // 更新图标
    this.updateIcon(theme);

    // 保存到localStorage
    localStorage.setItem('theme', theme);

    // 触发自定义事件（便于其他组件监听主题变化）
    window.dispatchEvent(new CustomEvent('themeChanged', { detail: theme }));
  }

  updateIcon(theme) {
    const isDark = theme === 'dark-theme';
    this.themeIcon.textContent = isDark ? '🌙' : '☀️';
    this.themeToggle.setAttribute('title', isDark ? '切换到亮色主题' : '切换到暗黑主题');
  }

  // 获取当前主题
  getCurrentTheme() {
    return this.body.classList.contains('dark-theme') ? 'dark-theme' : 'light-theme';
  }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
  window.themeManagerInstance = new ThemeManager(); // 保存实例

  window.addEventListener('themeChanged', function(event) { // 使用普通函数
    console.log('主题已切换至1:', event.detail);
    // window.themeManagerInstance.toggleTheme();
  });
});


// 用户下拉菜单功能
document.addEventListener('DOMContentLoaded', function() {
  const userDropdown = document.getElementById('userDropdown');
  const dropdownMenu = document.getElementById('dropdownMenu');

  if (userDropdown && dropdownMenu) {
    userDropdown.addEventListener('click', function(e) {
      e.stopPropagation();
      dropdownMenu.style.display = dropdownMenu.style.display === 'block' ? 'none' : 'block';
    });

    // 点击其他地方关闭下拉菜单
    document.addEventListener('click', function() {
      dropdownMenu.style.display = 'none';
    });
  }
});

// 滚动到顶部/底部功能
class ScrollManager {
  constructor() {
    this.scrollTopBtn = document.getElementById('scrollTopBtn');
    this.scrollBottomBtn = document.getElementById('scrollBottomBtn');
    this.scrollThreshold = 300; // 滚动超过300px时显示按钮

    this.init();
  }

  init() {
    if (this.scrollTopBtn && this.scrollBottomBtn) {
      // 绑定滚动事件
      window.addEventListener('scroll', () => this.handleScroll());

      // 绑定按钮点击事件
      this.scrollTopBtn.addEventListener('click', () => this.scrollToTop());
      this.scrollBottomBtn.addEventListener('click', () => this.scrollToBottom());

      // 初始检查
      this.handleScroll();
    }
  }

  handleScroll() {
    // 检查滚动位置
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
    const scrollHeight = document.documentElement.scrollHeight;
    const clientHeight = document.documentElement.clientHeight;
    const scrolledToBottom = scrollTop + clientHeight >= scrollHeight - 5;

    // 显示/隐藏回到顶部按钮
    if (scrollTop > this.scrollThreshold) {
      this.scrollTopBtn.classList.remove('hidden');
    } else {
      this.scrollTopBtn.classList.add('hidden');
    }

    // 显示/隐藏回到底部按钮
    if (!scrolledToBottom) {
      this.scrollBottomBtn.classList.remove('hidden');
    } else {
      this.scrollBottomBtn.classList.add('hidden');
    }
  }

  scrollToTop() {
    window.scrollTo({
      top: 0,
      behavior: 'smooth'
    });
  }

  scrollToBottom() {
    window.scrollTo({
      top: document.documentElement.scrollHeight,
      behavior: 'smooth'
    });
  }
}

// 页面加载完成后初始化滚动管理器
document.addEventListener('DOMContentLoaded', () => {
  // 初始化滚动管理器
  window.scrollManager = new ScrollManager();
});

// 搜索功能
document.addEventListener('DOMContentLoaded', function() {
  const searchInput = document.getElementById('searchInput');
  const searchDropdown = document.getElementById('searchDropdown');
  const searchOptions = document.querySelectorAll('.search-option');
  const searchButton = document.getElementById('searchButton');
  const searchTerms = document.querySelectorAll('.search-term');


  // 在搜索功能的DOMContentLoaded事件监听器中添加以下代码

// 添加键盘导航功能
  let selectedIndex = -1; // 当前选中的索引

// 键盘事件监听
  searchInput.addEventListener('keydown', function(e) {
    if (searchOptions.length === 0) return;

    switch(e.key) {
      case 'ArrowDown':
        e.preventDefault();
        selectedIndex = (selectedIndex + 1) % searchOptions.length;
        updateSelection();
        break;
      case 'ArrowUp':
        e.preventDefault();
        selectedIndex = (selectedIndex - 1 + searchOptions.length) % searchOptions.length;
        updateSelection();
        break;
      case 'Enter':
        e.preventDefault();
        if (selectedIndex >= 0 && selectedIndex < searchOptions.length) {
          // 触发选中项的点击事件
          searchOptions[selectedIndex].click();
        } else if (searchInput.value.trim() !== '') {
          // 默认搜索帖子
          performSearch(searchInput.value.trim(), 'posts');
          searchDropdown.style.display = 'none';
        }
        break;
      case 'Escape':
        searchDropdown.style.display = 'none';
        selectedIndex = -1;
        clearSelection();
        break;
    }
  });

// 更新选中状态
  function updateSelection() {
    clearSelection();
    if (selectedIndex >= 0 && selectedIndex < searchOptions.length) {
      searchOptions[selectedIndex].classList.add('selected');
      // 滚动到可视区域
      searchOptions[selectedIndex].scrollIntoView({ block: 'nearest' });
    }
  }

// 清除所有选中状态
  function clearSelection() {
    searchOptions.forEach(option => {
      option.classList.remove('selected');
    });
  }

// 点击选项时也要更新选中状态
  searchOptions.forEach((option, index) => {
    option.addEventListener('mouseenter', function() {
      selectedIndex = index;
      updateSelection();
    });

    option.addEventListener('click', function() {
      selectedIndex = index;
      updateSelection();
      // 原有的点击逻辑
      const searchType = this.getAttribute('data-type');
      const searchTerm = searchInput.value.trim();
      if (searchTerm !== '') {
        performSearch(searchTerm, searchType);
        searchDropdown.style.display = 'none';
      }
    });
  });

// 显示下拉菜单时默认选中第一个选项
  searchInput.addEventListener('focus', function() {
    updateSearchKeywords();
    updateSearchTerms();
    if (searchInput.value.trim() !== '') {
      searchDropdown.style.display = 'block';
      // 默认选中第一个选项
      selectedIndex = 0;
      updateSelection();
    }
  });

// 输入内容时控制下拉菜单显示并更新搜索词
  searchInput.addEventListener('input', function() {
    updateSearchKeywords();
    updateSearchTerms();
    if (searchInput.value.trim() !== '') {
      searchDropdown.style.display = 'block';
      // 重置选中状态
      selectedIndex = 0;
      updateSelection();
    } else {
      searchDropdown.style.display = 'none';
      selectedIndex = -1;
      clearSelection();
    }
  });


  // 输入框失去焦点时隐藏下拉菜单（延迟以允许点击选项）
  searchInput.addEventListener('blur', function() {
    setTimeout(() => {
      searchDropdown.style.display = 'none';
    }, 200);
  });

  function updateSearchKeywords() {
    const keyword = document.getElementById('searchInput').value;
    const keywords = document.querySelectorAll('.search-keyword');
    keywords.forEach(element => {
      element.textContent = keyword;
    });
  }


  // 更新搜索词显示
  function updateSearchTerms() {
    const searchTerm = searchInput.value.trim();
    searchTerms.forEach(term => {
      term.textContent = searchTerm;
    });
  }

  // 点击搜索选项
  searchOptions.forEach(option => {
    option.addEventListener('click', function() {
      const searchType = this.getAttribute('data-type');
      const searchTerm = searchInput.value.trim();

      if (searchTerm !== '') {
        performSearch(searchTerm, searchType);
        searchDropdown.style.display = 'none';
      }
    });
  });

  // 点击搜索按钮
// 点击搜索按钮 - 修改现有的事件处理函数
  searchButton.addEventListener('click', function() {
    // 如果有选中的搜索选项，使用该选项进行搜索
    if (selectedIndex >= 0 && selectedIndex < searchOptions.length) {
      const selectedOption = searchOptions[selectedIndex];
      const searchType = selectedOption.getAttribute('data-type');
      const searchTerm = searchInput.value.trim();
      if (searchTerm !== '') {
        performSearch(searchTerm, searchType);
        searchDropdown.style.display = 'none';
      }
    } else {
      // 默认搜索帖子
      const searchTerm = searchInput.value.trim();
      if (searchTerm !== '') {
        performSearch(searchTerm, 'posts');
      }
    }
  });


  // 回车键搜索
  searchInput.addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
      const searchTerm = searchInput.value.trim();
      if (searchTerm !== '') {
        performSearch(searchTerm, 'posts');
        searchDropdown.style.display = 'none';
      }
    }
  });

  // 执行搜索功能
  function performSearch(term, type) {
    switch(type) {
      case 'posts':
        // 搜索帖子逻辑
        window.location.href = `/search?q=${encodeURIComponent(term)}`;
        break;
      case 'users':
        // 搜索用户逻辑
        window.location.href = `/member?q=${encodeURIComponent(term)}`;
        break;
      case 'google':
        // 谷歌搜索逻辑
        window.open(`https://www.google.com/search?q=${encodeURIComponent(term)}`, '_blank');
        break;
    }
  }

  // 初始化时更新一次搜索词（如果有默认值）
  if (searchInput && searchInput.value) {
    updateSearchTerms();
  }
});

document.addEventListener('DOMContentLoaded', function() {
  // 获取菜单项和内容区域
  const menuItems = document.querySelectorAll('.operate-menu .menu-item');
  const tabContents = document.querySelectorAll('.tab-content');

  // 页面加载时根据URL参数激活对应tab
  activateTabFromUrl();

  // 为菜单项绑定点击事件
  menuItems.forEach(item => {
    item.addEventListener('click', function() {
      // 清除所有激活状态
      menuItems.forEach(menuItem => menuItem.classList.remove('is-active'));
      tabContents.forEach(content => content.classList.remove('active'));

      // 激活当前菜单项
      this.classList.add('is-active');

      // 显示对应的内容区域
      const tabName = this.getAttribute('data-tab');
      const targetTab = document.getElementById(tabName + '-tab');
      if (targetTab) {
        targetTab.classList.add('active');
      }

      // 更新URL参数
      updateUrlParameter('tab', tabName);

      // 重置page参数
      resetPageParameter();
    });
  });

  // 根据URL参数激活对应tab
  function activateTabFromUrl() {
    const urlParams = new URLSearchParams(window.location.search);
    const tabParam = urlParams.get('tab');

    if (tabParam) {
      // 移除所有激活状态
      menuItems.forEach(menuItem => menuItem.classList.remove('is-active'));
      tabContents.forEach(content => content.classList.remove('active'));

      // 激活对应tab
      const targetMenuItem = document.querySelector(`.menu-item[data-tab="${tabParam}"]`);
      const targetTab = document.getElementById(tabParam + '-tab');

      if (targetMenuItem && targetTab) {
        targetMenuItem.classList.add('is-active');
        targetTab.classList.add('active');
      }
    }
  }

  // 更新URL参数
  function updateUrlParameter(param, value) {
    const url = new URL(window.location);
    url.searchParams.set(param, value);
    window.history.replaceState({}, '', url);
  }

  // 重置页面参数函数
  function resetPageParameter() {
    const url = new URL(window.location);
    const currentPage = url.searchParams.get('page');

    // 如果当前有page参数且不为1，则移除page参数
    if (currentPage && currentPage !== '1') {
      url.searchParams.delete('page');
      window.history.replaceState({}, '', url);
    }
  }
});







