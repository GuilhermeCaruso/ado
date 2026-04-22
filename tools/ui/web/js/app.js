(() => {
  const tabs = document.querySelectorAll('.tab');
  const pages = document.querySelectorAll('.page');

  function activate(tabName) {
    tabs.forEach(t => t.classList.toggle('active', t.dataset.tab === tabName));
    pages.forEach(p => p.classList.toggle('active', p.id === tabName));
    history.replaceState(null, '', `#${tabName}`);
  }

  tabs.forEach(tab => tab.addEventListener('click', () => activate(tab.dataset.tab)));

  const initial = location.hash.slice(1);
  if (initial && document.getElementById(initial)) {
    activate(initial);
  }
})();
