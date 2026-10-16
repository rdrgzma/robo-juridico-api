// Estado Global da Aplicação Desktop
let estadoApp = {
  hwid: "",
  licencaAtiva: false,
  ultimoCalculo: null,
  ultimaConsulta: null,
};

// ==========================================================================
// Gerenciamento de Tema (Claro / Escuro)
// ==========================================================================
function obterTemaAtual() {
  return document.documentElement.getAttribute('data-theme') || 'dark';
}

function aplicarTema(tema) {
  document.documentElement.setAttribute('data-theme', tema);
  try {
    localStorage.setItem('robo_juridico_theme', tema);
  } catch (e) {}
  atualizarBotaoTema(tema);
}

function alternarTema() {
  const atual = obterTemaAtual();
  const novoTema = atual === 'dark' ? 'light' : 'dark';
  aplicarTema(novoTema);
}

function atualizarBotaoTema(tema) {
  const icon = document.getElementById('theme-icon');
  const label = document.getElementById('theme-label');
  if (icon && label) {
    if (tema === 'light') {
      icon.innerText = '☀️';
      label.innerText = 'Tema Claro';
    } else {
      icon.innerText = '🌙';
      label.innerText = 'Tema Escuro';
    }
  }
}

// Navegação entre abas
function navegarPara(secaoID) {
  document.querySelectorAll('.view-section').forEach(sec => sec.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(item => item.classList.remove('active'));

  const sec = document.getElementById('view-' + secaoID);
  if (sec) sec.classList.add('active');

  const links = document.querySelectorAll('.nav-item');
  links.forEach(l => {
    if (l.getAttribute('onclick') && l.getAttribute('onclick').includes(secaoID)) {
      l.classList.add('active');
    }
  });

  if (secaoID === 'indices') carregarIndices();
  if (secaoID === 'historico') carregarHistoricoLocal();
}

// 1. Carregar Metadados e Status do Sistema
async function carregarInfoSistema() {
  try {
    const res = await fetch('/api/sistema/info');
    const data = await res.json();

    estadoApp.hwid = data.hardware_id;
    estadoApp.licencaAtiva = data.licenca_ativa;

    // Atualiza badges
    const badge = document.getElementById('badge-licenca-sidebar');
    const statusVal = document.getElementById('dash-lic-status');
    const statusVenc = document.getElementById('dash-lic-venc');

    if (data.licenca_ativa) {
      badge.className = 'badge badge-success';
      badge.innerText = 'LICENÇA ATIVA';
      statusVal.innerText = 'Ativa (' + (data.licenca_detalhes?.plano || 'Pericial Pro') + ')';
      statusVenc.innerText = data.licenca_detalhes?.valida_ate ? 'Vence em: ' + new Date(data.licenca_detalhes.valida_ate).toLocaleDateString('pt-BR') : 'Uso Ilimitado';
    } else {
      badge.className = 'badge badge-warning';
      badge.innerText = 'MODO TRIAL / DEMO';
      statusVal.innerText = 'Demonstração Ativa';
      statusVenc.innerText = 'Ativação pendente';
    }

    document.getElementById('lbl-hwid-curto').innerText = data.hardware_id;
    document.getElementById('dash-hwid').innerText = data.hardware_id;
    document.getElementById('txt-hwid-display').value = data.hardware_id;
    document.getElementById('dash-total-indices').innerText = data.total_indices || 27;

  } catch (err) {
    console.error("Erro ao carregar status do sistema:", err);
  }
}

// 2. Validação e Máscara Dinâmica de CNJ
async function formatarEValidarCNJ(valor) {
  let digitos = valor.replace(/\D/g, '');
  if (digitos.length > 20) digitos = digitos.substring(0, 20);

  // Aplica máscara NNNNNNN-DD.AAAA.J.TR.OOOO
  let formatado = digitos;
  if (digitos.length > 7) formatado = digitos.substring(0, 7) + '-' + digitos.substring(7);
  if (digitos.length > 9) formatado = formatado.substring(0, 10) + '.' + formatado.substring(10);
  if (digitos.length > 13) formatado = formatado.substring(0, 15) + '.' + formatado.substring(15);
  if (digitos.length > 14) formatado = formatado.substring(0, 17) + '.' + formatado.substring(17);
  if (digitos.length > 16) formatado = formatado.substring(0, 20) + '.' + formatado.substring(20);

  document.getElementById('txt-numero-cnj').value = formatado;

  const badge = document.getElementById('badge-validacao-cnj');
  if (digitos.length === 20) {
    try {
      const res = await fetch('/api/processo/validar-cnj', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ numero: formatado })
      });
      const data = await res.json();
      badge.style.display = 'block';
      if (data.valido) {
        badge.style.color = 'var(--success)';
        badge.innerText = '✔ Módulo 97 Válido (' + (data.tribunal || 'Tribunal Identificado') + ')';
      } else {
        badge.style.color = 'var(--danger)';
        badge.innerText = '✖ Dígito verificador CNJ inválido';
      }
    } catch (e) {
      badge.style.display = 'none';
    }
  } else {
    badge.style.display = 'none';
  }
}

// 3. Consulta de Processo (DataJud + DJEN)
async function executarConsultaProcesso() {
  const cnj = document.getElementById('txt-numero-cnj').value.trim();
  const modo = document.getElementById('sel-modo-consulta').value;

  if (!cnj) {
    alert("Por favor, digite o número do processo.");
    return;
  }

  const box = document.getElementById('resultado-consulta-box');
  box.style.display = 'block';
  document.getElementById('res-proc-titulo').innerText = 'Consultando fontes oficiais...';

  try {
    const res = await fetch('/api/processo/consultar', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ numero: cnj, modo: modo })
    });
    const data = await res.json();

    if (!data.sucesso || !data.processo) {
      document.getElementById('res-proc-titulo').innerText = 'Nenhum registro encontrado nas fontes públicas';
      document.getElementById('res-proc-tribunal').innerText = data.erro || 'Verifique se o processo está em segredo de justiça.';
      return;
    }

    const p = data.processo;
    estadoApp.ultimaConsulta = p;

    document.getElementById('res-proc-titulo').innerText = p.numero || cnj;
    document.getElementById('res-proc-tribunal').innerText = 'Tribunal: ' + (p.tribunal || 'DataJud/DJEN') + ' · Atualizado em: ' + new Date().toLocaleTimeString('pt-BR');
    document.getElementById('res-proc-classe').innerText = p.classe || 'Não especificada';
    document.getElementById('res-proc-assunto').innerText = p.assunto || 'Cível Geral';
    document.getElementById('res-proc-orgao').innerText = p.orgao_julgador || 'Vara Cível';

    const movBox = document.getElementById('res-proc-movimentacoes');
    movBox.innerHTML = '';

    const movs = p.movimentacoes || [];
    if (movs.length === 0) {
      movBox.innerHTML = '<div class="text-muted" style="font-size: 13px;">Sem movimentações registradas recentemente.</div>';
    } else {
      movs.slice(0, 15).forEach(m => {
        const item = document.createElement('div');
        item.style = "padding: 8px 12px; background: var(--bg-subtle); border-radius: 4px; border-left: 3px solid var(--accent); border: 1px solid var(--border);";
        item.innerHTML = `
          <div class="text-muted" style="font-size: 11px;">${m.data || 'Data N/D'}</div>
          <div style="font-size: 13px; font-weight: 600; color: var(--text-main); margin-top: 2px;">${m.nome || m.descricao || 'Movimentação'}</div>
        `;
        movBox.appendChild(item);
      });
    }

  } catch (err) {
    alert("Erro ao consultar processo: " + err.message);
  }
}

// 4. Liquidação Judicial de Sentença
function adicionarLinhaParcela() {
  const tbody = document.querySelector('#tbl-parcelas-input tbody');
  const tr = document.createElement('tr');
  tr.innerHTML = `
    <td><input type="text" value="Parcela ${tbody.children.length + 1}" style="width: 100%;"></td>
    <td><input type="date" value="2023-01-10"></td>
    <td><input type="number" step="0.01" value="1000.00" style="width: 100%;"></td>
    <td>
      <select style="width: 100%;">
        <option value="DEBITO">Débito</option>
        <option value="CUSTA_PROCESSUAL">Custas Forenses</option>
        <option value="PAGAMENTO_PARCIAL">Abatimento</option>
      </select>
    </td>
    <td><button class="btn btn-secondary btn-sm" onclick="removerLinhaParcela(this)">❌</button></td>
  `;
  tbody.appendChild(tr);
}

function removerLinhaParcela(btn) {
  const tbody = document.querySelector('#tbl-parcelas-input tbody');
  if (tbody.children.length > 1) {
    btn.closest('tr').remove();
  } else {
    alert("Deve haver pelo menos uma parcela no cálculo.");
  }
}

async function executarCalculoLiquidacao() {
  const dataAdQuem = document.getElementById('calc-data-ad-quem').value;
  const indexador = document.getElementById('calc-indexador').value;
  const tipoJuros = document.getElementById('calc-tipo-juros').value;
  const art354 = document.getElementById('chk-art-354').checked;
  const multa523 = document.getElementById('chk-art-523-multa').checked;
  const honorarios523 = document.getElementById('chk-art-523-honorarios').checked;

  const rows = document.querySelectorAll('#tbl-parcelas-input tbody tr');
  const parcelas = [];

  rows.forEach((tr, idx) => {
    const inputs = tr.querySelectorAll('input');
    const select = tr.querySelector('select');
    parcelas.push({
      numero_item: idx + 1,
      descricao: inputs[0].value,
      data_vencimento: inputs[1].value,
      valor_original: parseFloat(inputs[2].value) || 0,
      tipo: select.value
    });
  });

  const payload = {
    data_atualizacao: dataAdQuem,
    indexador_codigo: indexador,
    imputar_art_354_cc: art354,
    juros_mora: {
      tipo: tipoJuros,
      percentual_mensal: 1.0,
      modo: "SIMPLES"
    },
    fase_execucao: {
      aplicar_art_523_multa: multa523,
      aplicar_art_523_honorarios: honorarios523
    },
    parcelas: parcelas
  };

  try {
    const res = await fetch('/api/calculo/liquidacao', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const data = await res.json();

    if (!data.sucesso && !data.ok) {
      alert("Erro no cálculo: " + (data.erro || data.mensagem));
      return;
    }

    estadoApp.ultimoCalculo = data;
    exibirResultadoLiquidacao(data);

  } catch (err) {
    alert("Erro na requisição: " + err.message);
  }
}

function exibirResultadoLiquidacao(data) {
  const box = document.getElementById('resultado-liquidacao-box');
  box.style.display = 'block';

  const r = data.resumo;
  const fmt = (v) => 'R$ ' + (parseFloat(v) || 0).toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

  document.getElementById('res-total-original').innerText = fmt(r.total_principal_original);
  document.getElementById('res-total-corrigido').innerText = fmt(r.total_principal_corrigido);
  document.getElementById('res-total-juros').innerText = fmt(r.total_juros_mora);
  document.getElementById('res-total-geral').innerText = fmt(r.total_geral_devido);

  const tbody = document.querySelector('#tbl-memoria-analitica tbody');
  tbody.innerHTML = '';

  const itens = data.itens_liquidados || data.itens || [];
  itens.forEach(it => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>${it.numero_item}</td>
      <td><strong>${it.descricao}</strong></td>
      <td>${it.data_vencimento}</td>
      <td>${fmt(it.valor_original)}</td>
      <td>${parseFloat(it.fator_correcao).toFixed(6)}</td>
      <td>${fmt(it.valor_corrigido)}</td>
      <td>${fmt(it.valor_juros_mora)}</td>
      <td><strong class="text-accent">${fmt(it.total_item)}</strong></td>
    `;
    tbody.appendChild(tr);
  });

  box.scrollIntoView({ behavior: 'smooth' });
}

// 5. Exportar CSV
async function exportarCSVMemoria() {
  if (!estadoApp.ultimoCalculo) {
    alert("Nenhum cálculo processado para exportar.");
    return;
  }
  try {
    const res = await fetch('/api/calculo/exportar-csv', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(estadoApp.ultimoCalculo)
    });
    const blob = await res.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = "laudo_calculo_judicial.csv";
    document.body.appendChild(a);
    a.click();
    a.remove();
  } catch (err) {
    alert("Erro ao exportar CSV: " + err.message);
  }
}

// 6. Revisão Bancária
async function executarCalculoRevisional() {
  const valorFinanciado = parseFloat(document.getElementById('rev-valor-financiado').value) || 0;
  const numParcelas = parseInt(document.getElementById('rev-num-parcelas').value) || 0;
  const taxaContrato = parseFloat(document.getElementById('rev-taxa-contrato').value) || 0;
  const taxaMercado = parseFloat(document.getElementById('rev-taxa-mercado').value) || 0;
  const sistema = document.getElementById('rev-sistema').value;
  const restituicaoDobro = document.getElementById('chk-restituicao-dobro').checked;

  const payload = {
    valor_financiado: valorFinanciado,
    numero_parcelas: numParcelas,
    taxa_contratada_mensal: taxaContrato,
    taxa_mercado_mensal: taxaMercado,
    sistema_amortizacao: sistema,
    restituicao_em_dobro: restituicaoDobro
  };

  try {
    const res = await fetch('/api/calculo/revisional', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const data = await res.json();

    if (!res.ok) {
      alert("Erro na revisão: " + (data.erro || 'Falha no recálculo'));
      return;
    }

    const fmt = (v) => 'R$ ' + (parseFloat(v) || 0).toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    document.getElementById('resultado-revisional-box').style.display = 'block';
    document.getElementById('rev-pmt-contratada').innerText = fmt(data.prestacao_contratada);
    document.getElementById('rev-pmt-revisada').innerText = fmt(data.prestacao_revisada_inicial);
    
    const diff = (parseFloat(data.prestacao_contratada) - parseFloat(data.prestacao_revisada_inicial)) || 0;
    document.getElementById('rev-eco-parcela').innerText = fmt(diff);
    document.getElementById('rev-total-restituir').innerText = fmt(data.indebito_cobrado || data.diferenca_total_simples);

  } catch (err) {
    alert("Erro ao calcular revisional: " + err.message);
  }
}

// 7. Prazos Forenses
async function executarCalculoPrazo() {
  const dtDisp = document.getElementById('prazo-data-disp').value;
  const dias = parseInt(document.getElementById('prazo-dias').value) || 0;
  const rito = document.getElementById('prazo-rito').value;
  const tipoInt = document.getElementById('prazo-tipo-int').value;

  const payload = {
    data_disponibilizacao: dtDisp,
    dias_prazo: dias,
    rito: rito,
    tipo_intimacao: tipoInt
  };

  try {
    const res = await fetch('/api/calculo/prazos', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const data = await res.json();

    if (!data.sucesso) {
      alert("Erro na contagem de prazo: " + (data.erro || 'Falha'));
      return;
    }

    document.getElementById('resultado-prazo-box').style.display = 'block';
    document.getElementById('prazo-inicio-contagem').innerText = data.data_inicio_contagem;
    document.getElementById('prazo-vencimento-fatal').innerText = data.data_vencimento_fatal;

    const tbody = document.querySelector('#tbl-memoria-prazo tbody');
    tbody.innerHTML = '';
    (data.memoria_contagem || []).forEach(m => {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td>${m.data}</td>
        <td>${m.dia_semana}</td>
        <td><span class="badge ${m.util ? 'badge-success' : 'badge-warning'}">${m.util ? 'Útil' : 'Suspenso'}</span></td>
        <td><strong>${m.dia_contado > 0 ? m.dia_contado + 'º dia' : '-'}</strong></td>
        <td>${m.motivo || 'Expediente Forense Normal'}</td>
      `;
      tbody.appendChild(tr);
    });

  } catch (err) {
    alert("Erro ao calcular prazo: " + err.message);
  }
}

// 8. Índices Econômicos
async function carregarIndices() {
  try {
    const res = await fetch('/api/indices');
    const indices = await res.json();
    const tbody = document.querySelector('#tbl-indices-catalogo tbody');
    tbody.innerHTML = '';

    indices.forEach(idx => {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td><strong class="text-accent">${idx.codigo}</strong></td>
        <td>${idx.nome}</td>
        <td>${idx.fonte}</td>
        <td>${idx.periodicidade}</td>
        <td>${idx.data_inicio_disponivel || '2020-01-01'} a ${idx.data_fim_disponivel || 'Atual'}</td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {
    console.error("Erro ao listar índices:", err);
  }
}

async function sincronizarIndicesBacen() {
  try {
    const res = await fetch('/api/indices/sincronizar', { method: 'POST' });
    const data = await res.json();
    alert(data.mensagem || "Sincronização com Bacen e IBGE finalizada!");
    carregarIndices();
  } catch (err) {
    alert("Falha na sincronização: " + err.message);
  }
}

// 9. Histórico Local
async function carregarHistoricoLocal() {
  try {
    const res = await fetch('/api/historico');
    const lista = await res.json();
    const tbody = document.querySelector('#tbl-historico-lista tbody');
    tbody.innerHTML = '';

    if (lista.length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" style="text-align: center; color: var(--text-muted);">Nenhum cálculo salvo ainda.</td></tr>';
      return;
    }

    lista.forEach(item => {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td>${new Date(item.criado_em).toLocaleString('pt-BR')}</td>
        <td><span class="badge badge-success">${item.tipo}</span></td>
        <td><strong>${item.titulo}</strong></td>
        <td>${item.descricao || '-'}</td>
        <td><button class="btn btn-secondary btn-sm" onclick="excluirItemHistorico(${item.id})">Excluir</button></td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {
    console.error("Erro ao carregar histórico:", err);
  }
}

async function excluirItemHistorico(id) {
  if (!confirm("Deseja realmente remover este item do histórico local?")) return;
  try {
    await fetch('/api/historico/' + id, { method: 'DELETE' });
    carregarHistoricoLocal();
  } catch (e) {
    alert("Erro ao excluir.");
  }
}

// 10. Licenciamento
function copiarHWID() {
  const hwid = document.getElementById('txt-hwid-display').value;
  navigator.clipboard.writeText(hwid);
  alert("Machine ID copiado para a área de transferência: " + hwid);
}

async function ativarLicencaToken() {
  const token = document.getElementById('txt-token-licenca').value.trim();
  if (!token) {
    alert("Por favor, cole o token da licença.");
    return;
  }
  try {
    const res = await fetch('/api/licenca/ativar', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token: token })
    });
    const data = await res.json();
    if (data.sucesso) {
      alert("Licença ativada com sucesso para: " + data.licenca.nome_titular);
      carregarInfoSistema();
    } else {
      alert("Falha ao ativar: " + (data.erro || "Token inválido"));
    }
  } catch (err) {
    alert("Erro de conexão: " + err.message);
  }
}

// Inicialização
document.addEventListener('DOMContentLoaded', () => {
  const temaSalvo = localStorage.getItem('robo_juridico_theme') || 
    (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark');
  aplicarTema(temaSalvo);
  carregarInfoSistema();
});
