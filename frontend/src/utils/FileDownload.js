export const saveAs = (filename, blob) => {
  const element = document.createElement('a');
  element.href = URL.createObjectURL(blob);
  element.download = filename;

  element.style.display = 'none';
  document.body.appendChild(element);

  element.click();

  document.body.removeChild(element);
};

export const saveText = (filename, content) => {
  const blob = new Blob([content], {
    type: 'text/plain;charset=utf-8',
  });

  saveAs(filename, blob);
};
