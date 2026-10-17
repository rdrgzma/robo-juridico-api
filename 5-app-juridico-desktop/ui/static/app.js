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
function limparResultadosConsulta() {
  const box = document.getElementById('resultado-consulta-box');
  if (box) {
    box.style.display = 'none';
  }
  const loadingBox = document.getElementById('consulta-loading-box');
  if (loadingBox) {
    loadingBox.style.display = 'none';
  }
  estadoApp.ultimaConsulta = null;
  estadoApp.todasMovimentacoes = [];
  estadoApp.mostrarTodasMovs = false;

  const titulo = document.getElementById('res-proc-titulo');
  if (titulo) titulo.innerText = 'Processo ---';
  const trib = document.getElementById('res-proc-tribunal');
  if (trib) trib.innerText = 'Tribunal: ---';
  const badgeGrau = document.getElementById('badge-proc-grau');
  if (badgeGrau) badgeGrau.style.display = 'none';
  const btnTrib = document.getElementById('btn-proc-link-tribunal');
  if (btnTrib) {
    btnTrib.style.display = 'none';
    btnTrib.href = '#';
  }
  const btnPeca = document.getElementById('btn-proc-link-peca');
  if (btnPeca) {
    btnPeca.style.display = 'none';
    btnPeca.href = '#';
  }

  // Limpa barra de fontes e avisos
  const fontesBar = document.getElementById('res-proc-fontes-bar');
  if (fontesBar) fontesBar.style.display = 'none';
  const fontesBadges = document.getElementById('res-proc-fontes-badges');
  if (fontesBadges) fontesBadges.innerHTML = '';
  const avisosLista = document.getElementById('res-proc-avisos-lista');
  if (avisosLista) {
    avisosLista.style.display = 'none';
    avisosLista.innerHTML = '';
  }

  // Limpa metadados e polos do processo anterior
  const elClasse = document.getElementById('res-proc-classe');
  if (elClasse) elClasse.innerText = '---';
  const elAssunto = document.getElementById('res-proc-assunto');
  if (elAssunto) elAssunto.innerText = '---';
  const elOrgao = document.getElementById('res-proc-orgao');
  if (elOrgao) elOrgao.innerText = '---';
  const elDatas = document.getElementById('res-proc-datas');
  if (elDatas) elDatas.innerText = '---';
  const elPoloAtivo = document.getElementById('res-proc-polo-ativo');
  if (elPoloAtivo) elPoloAtivo.innerText = '---';
  const elPoloPassivo = document.getElementById('res-proc-polo-passivo');
  if (elPoloPassivo) elPoloPassivo.innerText = '---';
  const elAdv = document.getElementById('res-proc-advogados');
  if (elAdv) elAdv.innerText = '---';
  const badgeTotalMovs = document.getElementById('badge-total-movs');
  if (badgeTotalMovs) badgeTotalMovs.innerText = '0 registros';
  const btnToggle = document.getElementById('btn-toggle-todas-movs');
  if (btnToggle) btnToggle.style.display = 'none';

  const movBox = document.getElementById('res-proc-movimentacoes');
  if (movBox) movBox.innerHTML = '';
}

async function formatarEValidarCNJ(valor) {
  // Limpa os resultados da consulta anterior sempre que o número for digitado ou alterado
  const box = document.getElementById('resultado-consulta-box');
  if (box && box.style.display !== 'none') {
    limparResultadosConsulta();
  }

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

// Utilitário de formatação de datas
function formatarDataHora(str) {
  if (!str) return 'N/D';
  if (str.length === 10 && str.includes('-')) {
    const parts = str.split('-');
    return `${parts[2]}/${parts[1]}/${parts[0]}`;
  }
  try {
    const d = new Date(str);
    if (!isNaN(d.getTime())) {
      return d.toLocaleDateString('pt-BR');
    }
  } catch (e) {}
  return str;
}

// 3. Consulta de Processo (DataJud + DJEN)
async function executarConsultaProcesso() {
  const cnj = document.getElementById('txt-numero-cnj').value.trim();
  const modo = document.getElementById('sel-modo-consulta').value;

  if (!cnj) {
    alert("Por favor, digite o número do processo.");
    return;
  }

  // 1. Limpa completamente os dados da consulta anterior
  limparResultadosConsulta();

  const btn = document.getElementById('btn-exec-consulta-processo');
  const btnOriginalHTML = btn ? btn.innerHTML : 'Buscar Processo';
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner"></span> Consultando...';
  }

  // 2. Exibe o card elegante de consulta às fontes oficiais
  const loadingBox = document.getElementById('consulta-loading-box');
  const loadingCNJ = document.getElementById('consulta-loading-cnj');
  const pillDataJud = document.getElementById('loading-pill-datajud');
  const pillDJEN = document.getElementById('loading-pill-djen');

  if (loadingCNJ) loadingCNJ.innerText = cnj;
  if (pillDataJud) {
    pillDataJud.style.display = modo === 'rapido' ? 'none' : 'inline-flex';
  }
  if (pillDJEN) {
    pillDJEN.style.display = 'inline-flex';
  }
  if (loadingBox) {
    loadingBox.style.display = 'flex';
  }

  const box = document.getElementById('resultado-consulta-box');

  try {
    const res = await fetch('/api/processo/consultar', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ numero: cnj, modo: modo })
    });
    const data = await res.json();

    if (loadingBox) loadingBox.style.display = 'none';
    if (box) box.style.display = 'block';

    if (!data.sucesso || !data.processo) {
      document.getElementById('res-proc-titulo').innerText = 'Nenhum registro encontrado nas fontes públicas';
      document.getElementById('res-proc-tribunal').innerText = data.erro || 'Verifique se o processo está em segredo de justiça ou se o número está correto.';
      return;
    }

    const p = data.processo;
    const avisos = data.avisos || [];
    estadoApp.ultimaConsulta = p;
    estadoApp.todasMovimentacoes = p.movimentacoes || [];
    estadoApp.mostrarTodasMovs = false;

    // Cabeçalho e Tribunal
    document.getElementById('res-proc-titulo').innerText = p.titulo || p.numero || cnj;
    document.getElementById('res-proc-tribunal').innerText = 'Tribunal: ' + (p.tribunal || 'DataJud/DJEN') + (p.segmento ? ' · ' + p.segmento : '') + ' · Consultado em: ' + new Date().toLocaleTimeString('pt-BR');

    // Renderiza Aviso de Fontes Consultadas e Alertas do Orquestrador
    const fontesBar = document.getElementById('res-proc-fontes-bar');
    const fontesBadges = document.getElementById('res-proc-fontes-badges');
    const avisosLista = document.getElementById('res-proc-avisos-lista');

    if (fontesBar && fontesBadges) {
      const fontesProc = Array.isArray(p.fontes) && p.fontes.length > 0 ? p.fontes : [];
      fontesBadges.innerHTML = '';

      if (fontesProc.includes('datajud')) {
        const qtdMovDJ = (p.movimentacoes || []).filter(m => m.fonte === 'datajud').length;
        const spanDJ = document.createElement('span');
        spanDJ.className = 'fonte-badge fonte-badge-datajud';
        spanDJ.innerHTML = `🏛️ DataJud / CNJ (${qtdMovDJ} andamento${qtdMovDJ === 1 ? '' : 's'})`;
        fontesBadges.appendChild(spanDJ);
      }

      if (fontesProc.includes('comunica')) {
        const qtdPubs = (p.publicacoes || []).length || (p.movimentacoes || []).filter(m => m.fonte === 'comunica').length;
        const spanCom = document.createElement('span');
        spanCom.className = 'fonte-badge fonte-badge-comunica';
        spanCom.innerHTML = `📰 DJEN / Comunica PJe (${qtdPubs} publicação${qtdPubs === 1 ? '' : 'ões'})`;
        fontesBadges.appendChild(spanCom);
      }

      if (fontesBadges.children.length === 0) {
        const spanGen = document.createElement('span');
        spanGen.className = 'fonte-badge fonte-badge-datajud';
        spanGen.innerText = '🏛️ Base Oficial CNJ';
        fontesBadges.appendChild(spanGen);
      }

      if (avisosLista) {
        if (avisos.length > 0) {
          avisosLista.innerHTML = '';
          avisos.forEach(av => {
            const divAviso = document.createElement('div');
            divAviso.className = 'aviso-item';
            divAviso.innerHTML = `<span>⚠️</span> <span>${av}</span>`;
            avisosLista.appendChild(divAviso);
          });
          avisosLista.style.display = 'flex';
        } else {
          avisosLista.style.display = 'none';
          avisosLista.innerHTML = '';
        }
      }

      fontesBar.style.display = 'flex';
    }

    // Badge de Grau
    const badgeGrau = document.getElementById('badge-proc-grau');
    if (badgeGrau) {
      if (p.grau) {
        badgeGrau.style.display = 'inline-block';
        badgeGrau.innerText = p.grau === 'TR' ? 'Turma Recursal (TR)' : (p.grau === 'G2' ? '2º Grau (TJ)' : '1º Grau (Vara/JESP)');
      } else {
        badgeGrau.style.display = 'none';
      }
    }

    // Botão de Link Externo (Autos no Tribunal)
    const btnLink = document.getElementById('btn-proc-link-tribunal');
    if (btnLink) {
      if (p.link) {
        btnLink.href = p.link;
        btnLink.style.display = 'inline-flex';
      } else {
        btnLink.style.display = 'none';
      }
    }

    // Botão de Peça Recente (DJEN/PJe)
    const btnPeca = document.getElementById('btn-proc-link-peca');
    if (btnPeca) {
      const pecaUrl = p.link_peca || (p.link && (p.link.includes('ConsultaDocumento') || p.link.includes('listView.seam?x=')) ? p.link : null);
      if (pecaUrl) {
        btnPeca.href = pecaUrl;
        btnPeca.style.display = 'inline-flex';
      } else {
        btnPeca.style.display = 'none';
      }
    }

    // Metadados
    document.getElementById('res-proc-classe').innerText = p.classe || 'Não especificada';
    
    // Tratamento de Assunto Inteligente (Evita duplicar classe)
    let assuntoExibicao = p.assunto || 'Cível Geral';
    if (p.assunto && p.classe && p.assunto.trim().toLowerCase() === p.classe.trim().toLowerCase()) {
      assuntoExibicao = 'Cível Geral (Não especificado)';
    }
    document.getElementById('res-proc-assunto').innerText = assuntoExibicao;

    const orgaoTexto = (p.orgao_julgador || 'Vara Cível') + (p.comarca ? ' (' + p.comarca + ')' : '');
    document.getElementById('res-proc-orgao').innerText = orgaoTexto;

    // Datas Processuais
    const dataAjuiz = p.data_ajuizamento_formatada || formatarDataHora(p.data_ajuizamento) || 'N/D';
    const dataUltima = p.data_ultima_movimentacao_formatada || formatarDataHora(p.data_ultima_movimentacao) || 'N/D';
    document.getElementById('res-proc-datas').innerText = `Ajuizamento: ${dataAjuiz} · Última: ${dataUltima}`;

    // Polos (Partes) e Advogados
    const poloAtivoBox = document.getElementById('res-proc-polo-ativo');
    if (poloAtivoBox) {
      const autores = (p.polos?.ativo || []).map(a => a.nome).filter(Boolean);
      poloAtivoBox.innerText = autores.length > 0 ? autores.join('; ') : 'Não informado';
    }

    const poloPassivoBox = document.getElementById('res-proc-polo-passivo');
    if (poloPassivoBox) {
      const reus = (p.polos?.passivo || []).map(r => r.nome).filter(Boolean);
      poloPassivoBox.innerText = reus.length > 0 ? reus.join('; ') : 'Não informado';
    }

    const advBox = document.getElementById('res-proc-advogados');
    if (advBox) {
      const advs = (p.advogados || []).map(adv => `${adv.nome} (${adv.oab || 'OAB N/D'})`).filter(Boolean);
      advBox.innerText = advs.length > 0 ? advs.join(' · ') : 'Nenhum advogado cadastrado';
    }

    // Renderiza linha do tempo com badges
    renderizarMovimentacoes();

  } catch (err) {
    if (loadingBox) loadingBox.style.display = 'none';
    alert("Erro ao consultar processo: " + err.message);
  } finally {
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = btnOriginalHTML;
    }
  }
}

// Renderização detalhada da timeline de movimentações
function renderizarMovimentacoes() {
  const movBox = document.getElementById('res-proc-movimentacoes');
  if (!movBox) return;
  movBox.innerHTML = '';

  const movs = estadoApp.todasMovimentacoes || [];
  const total = movs.length;
  
  const badgeTotal = document.getElementById('badge-total-movs');
  if (badgeTotal) badgeTotal.innerText = `${total} registro(s) encontrado(s)`;

  const btnToggle = document.getElementById('btn-toggle-todas-movs');
  if (btnToggle) {
    if (total <= 15) {
      btnToggle.style.display = 'none';
    } else {
      btnToggle.style.display = 'inline-flex';
      btnToggle.innerText = estadoApp.mostrarTodasMovs ? 'Ver menos (15)' : `Ver todas (${total})`;
    }
  }

  const exibicao = estadoApp.mostrarTodasMovs ? movs : movs.slice(0, 15);

  if (exibicao.length === 0) {
    movBox.innerHTML = '<div class="text-muted" style="font-size: 13px; padding: 12px;">Sem movimentações registradas recentemente.</div>';
    return;
  }

  exibicao.forEach((m, idx) => {
    const item = document.createElement('div');
    const isDJEN = m.fonte === 'comunica' || m.tipo === 'intimacao';
    const borderCor = isDJEN ? '#a855f7' : 'var(--accent)';
    const badgeFonte = isDJEN 
      ? '<span class="badge" style="background: rgba(168, 85, 247, 0.15); color: #a855f7; font-size: 10px; font-weight: 700;">DJEN / COMUNICA</span>' 
      : '<span class="badge" style="background: rgba(59, 130, 246, 0.15); color: var(--accent-text); font-size: 10px; font-weight: 700;">DATAJUD</span>';
    
    let badgeTipo = '';
    if (m.tipo === 'sentenca') {
      badgeTipo = '<span class="badge" style="background: rgba(239, 68, 68, 0.15); color: var(--danger); font-size: 10px; font-weight: 700;">SENTENÇA</span>';
    } else if (m.tipo === 'decisao') {
      badgeTipo = '<span class="badge" style="background: rgba(245, 158, 11, 0.15); color: var(--warning); font-size: 10px; font-weight: 700;">DECISÃO</span>';
    } else if (m.tipo === 'despacho') {
      badgeTipo = '<span class="badge" style="background: rgba(16, 185, 129, 0.15); color: var(--success); font-size: 10px; font-weight: 700;">DESPACHO</span>';
    }

    const dataFormatada = m.data_formatada || formatarDataHora(m.data);
    const urlPeca = m.link_peca || (m.link && (m.link.includes('ConsultaDocumento') || m.link.includes('listView.seam?x=') || m.link.includes('certidao') || m.link.includes('documento')) ? m.link : null);

    let linkBtn = '<div style="margin-left: auto; display: inline-flex; align-items: center; gap: 6px;">';
    if (urlPeca) {
      linkBtn += `<a href="${urlPeca}" target="_blank" rel="noopener noreferrer" style="color: #9333ea; background: rgba(147, 51, 234, 0.1); border: 1px solid rgba(147, 51, 234, 0.3); padding: 3px 9px; border-radius: 4px; font-size: 11px; font-weight: 700; text-decoration: none; display: inline-flex; align-items: center; gap: 4px;">📄 Abrir Peça (PDF) ↗</a>`;
    }
    linkBtn += `<button type="button" onclick="abrirDocumentoMovimentacao(${idx})" style="color: var(--accent-text); background: rgba(59, 130, 246, 0.1); border: 1px solid rgba(59, 130, 246, 0.3); padding: 3px 9px; border-radius: 4px; font-size: 11px; font-weight: 700; cursor: pointer; display: inline-flex; align-items: center; gap: 4px;">📄 Ver Documento</button>`;
    linkBtn += '</div>';

    item.style = `padding: 10px 14px; background: var(--bg-subtle); border-radius: 6px; border-left: 4px solid ${borderCor}; border: 1px solid var(--border); display: flex; flex-direction: column; gap: 4px;`;
    item.innerHTML = `
      <div style="display: flex; align-items: center; gap: 8px; flex-wrap: wrap;">
        ${badgeFonte}
        ${badgeTipo}
        <span class="text-muted" style="font-size: 11px; font-weight: 600;">${dataFormatada}</span>
        ${m.orgao ? `<span class="text-muted" style="font-size: 11px;">· ${m.orgao}</span>` : ''}
        ${linkBtn}
      </div>
      <div style="font-size: 13px; font-weight: 600; color: var(--text-main); margin-top: 2px;">
        ${m.descricao || m.nome || 'Movimentação Processual'}
      </div>
      ${m.complemento && m.complemento !== m.descricao ? `<div style="font-size: 12px; color: var(--text-muted); line-height: 1.4; margin-top: 2px;">${m.complemento}</div>` : ''}
    `;
    movBox.appendChild(item);
  });
}

// Abre o Modal de Certidão / Documento da Movimentação (DataJud ou DJEN)
function abrirDocumentoMovimentacao(idx) {
  const movs = estadoApp.todasMovimentacoes || [];
  const m = movs[idx];
  const p = estadoApp.ultimaConsulta;
  if (!m || !p) return;

  const isDJEN = m.fonte === 'comunica';
  const dataFormatada = m.data_formatada || formatarDataHora(m.data);
  const urlPeca = m.link_peca || (m.link && (m.link.includes('ConsultaDocumento') || m.link.includes('listView.seam?x=') || m.link.includes('certidao') || m.link.includes('documento')) ? m.link : null);
  const urlTribunal = p.link || m.link || '#';

  document.getElementById('doc-modal-header-title').innerText = isDJEN
    ? 'Certidão de Publicação / Comunicação Judicial (DJEN)'
    : 'Certidão Oficial de Movimentação Processual (DataJud)';
  document.getElementById('doc-modal-header-sub').innerText = `Processo ${p.numero} · ${p.tribunal || 'CNJ'} · ${dataFormatada}`;

  document.getElementById('doc-paper-sistema').innerText = isDJEN
    ? 'DIÁRIO DE JUSTIÇA ELETRÔNICO NACIONAL — DJEN / COMUNICA PJE'
    : 'EXTRATO OFICIAL DE MOVIMENTAÇÃO PROCESSUAL — DATAJUD / CNJ';
  document.getElementById('doc-paper-tribunal').innerText = `${p.tribunal || 'TRIBUNAL'} · ${(p.segmento || 'Poder Judiciário').toUpperCase()}`;

  document.getElementById('doc-paper-numero').innerText = p.numero || '---';
  document.getElementById('doc-paper-classe').innerText = `${p.classe || 'Não especificada'}${p.assunto && p.assunto !== p.classe ? ' — ' + p.assunto : ''}`;
  document.getElementById('doc-paper-orgao').innerText = m.orgao || p.orgao_julgador || '---';
  document.getElementById('doc-paper-data').innerText = `${dataFormatada}${m.data && m.data.includes('T') ? ' (' + m.data.substring(11, 19) + ' UTC)' : ''}`;
  document.getElementById('doc-paper-partes').innerText = p.titulo || 'Não informado';

  const tipoMap = {
    sentenca: 'SENTENÇA / JULGAMENTO',
    decisao: 'DECISÃO INTERLOCUTÓRIA',
    despacho: 'DESPACHO JUDICIAL',
    intimacao: 'INTIMAÇÃO / COMUNICAÇÃO',
    publicacao_dje: 'PUBLICAÇÃO NO DIÁRIO ELETRÔNICO',
    audiencia: 'AUDIÊNCIA',
    outro: isDJEN ? 'COMUNICAÇÃO OFICIAL DJEN' : 'MOVIMENTAÇÃO PROCESSUAL (TPU/CNJ)'
  };
  document.getElementById('doc-paper-tipo-badge').innerText = tipoMap[m.tipo] || 'ANDAMENTO PROCESSUAL';
  document.getElementById('doc-paper-id-externo').innerText = `ID: ${m.id_externo || 'CNJ-REG'}`;

  document.getElementById('doc-paper-descricao').innerText = m.descricao || m.nome || 'Movimentação Processual';

  // Monta o teor completo do documento
  let teorTexto = m.complemento || m.descricao || '';
  if (!isDJEN && (!m.complemento || m.complemento === m.descricao)) {
    // Tenta localizar se há alguma publicação DJEN próxima ou exibe o registro estruturado completo do DataJud
    teorTexto = `REGISTRO OFICIAL DE ANDAMENTO NOS AUTOS\n\n` +
      `• Ato / Movimento (TPU/CNJ): ${m.descricao || '---'}\n` +
      `• Complemento Tabelado: ${m.complemento || m.descricao || '---'}\n` +
      `• Data e Hora do Lançamento: ${dataFormatada}${m.data && m.data.includes('T') ? ' às ' + m.data.substring(11, 19) : ''}\n` +
      `• Unidade Judiciária: ${m.orgao || p.orgao_julgador || '---'}${p.comarca ? ' (' + p.comarca + ')' : ''}\n` +
      `• Identificador de Sincronização DataJud: ${m.id_externo || 'N/D'}`;
  }
  document.getElementById('doc-paper-complemento').innerText = teorTexto;

  document.getElementById('doc-paper-emissao').innerText = new Date().toLocaleString('pt-BR');
  document.getElementById('doc-paper-hash').innerText = `FONTE: ${(m.fonte || 'DATAJUD').toUpperCase()} | REF: ${m.id_externo || p.numero_digitos}`;

  const btnPDF = document.getElementById('doc-modal-btn-pdf');
  if (btnPDF) {
    if (urlPeca) {
      btnPDF.href = urlPeca;
      btnPDF.style.display = 'inline-flex';
    } else {
      btnPDF.style.display = 'none';
    }
  }

  const btnTrib = document.getElementById('doc-modal-btn-tribunal');
  if (btnTrib) {
    if (urlTribunal && urlTribunal !== '#') {
      btnTrib.href = urlTribunal;
      btnTrib.style.display = 'inline-flex';
    } else {
      btnTrib.style.display = 'none';
    }
  }

  const modal = document.getElementById('modal-doc-mov');
  if (modal) {
    modal.style.display = 'flex';
  }
}

function fecharDocumentoMovimentacao() {
  const modal = document.getElementById('modal-doc-mov');
  if (modal) {
    modal.style.display = 'none';
  }
}

function imprimirDocumentoMovimentacao() {
  const conteudo = document.getElementById('doc-modal-printable');
  if (!conteudo) return;
  const win = window.open('', '_blank', 'width=860,height=720');
  if (!win) {
    window.print();
    return;
  }
  win.document.write(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <title>Certidão de Movimentação Processual</title>
  <style>
    body { font-family: "Segoe UI", Roboto, Georgia, serif; color: #0f172a; margin: 0; padding: 28px; background: #fff; }
    .doc-paper { max-width: 760px; margin: 0 auto; }
    .doc-paper-header { display: flex; align-items: center; gap: 16px; border-bottom: 2px solid #1e293b; padding-bottom: 14px; margin-bottom: 20px; }
    .doc-paper-brasao { font-size: 32px; }
    .doc-paper-rep { font-size: 11px; font-weight: 800; color: #475569; letter-spacing: 0.8px; }
    .doc-paper-sis { font-size: 14px; font-weight: 800; color: #0f172a; margin-top: 2px; }
    .doc-paper-trib { font-size: 12px; font-weight: 600; color: #2563eb; margin-top: 2px; }
    .doc-paper-meta-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px 18px; background: #f8fafc; border: 1px solid #cbd5e1; border-radius: 6px; padding: 14px 18px; margin-bottom: 20px; }
    .doc-meta-label { display: block; font-size: 10px; font-weight: 700; color: #64748b; }
    .doc-meta-value { font-size: 13px; font-weight: 700; color: #0f172a; }
    .doc-paper-section { border: 1px solid #cbd5e1; border-radius: 6px; padding: 20px; margin-bottom: 18px; }
    .doc-section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
    .doc-badge { padding: 3px 8px; font-size: 11px; font-weight: 800; background: #dbeafe; color: #1d4ed8; border-radius: 4px; }
    .doc-paper-mov-title { font-size: 16px; font-weight: 800; margin: 0 0 12px 0; padding-bottom: 8px; border-bottom: 1px dashed #cbd5e1; }
    .doc-paper-teor { font-size: 13px; line-height: 1.65; white-space: pre-wrap; word-break: break-word; }
    .doc-paper-notice { font-size: 11px; color: #475569; background: #f1f5f9; border-left: 4px solid #3b82f6; padding: 8px 12px; margin-bottom: 18px; }
    .doc-paper-footer { display: flex; justify-content: space-between; border-top: 1px solid #cbd5e1; padding-top: 10px; font-size: 10px; color: #64748b; }
  </style>
</head>
<body>
  ${conteudo.innerHTML}
  <script>window.onload = function() { window.print(); };<\/script>
</body>
</html>`);
  win.document.close();
}

// Alternar entre ver 15 ou todas as movimentações
function alternarTodasMovimentacoes() {
  estadoApp.mostrarTodasMovs = !estadoApp.mostrarTodasMovs;
  renderizarMovimentacoes();
}

// 1-Click Auto-Fill para Liquidação Judicial
function copiarDadosParaLiquidacao() {
  if (!estadoApp.ultimaConsulta) {
    alert("Nenhum processo consultado no momento.");
    return;
  }
  const p = estadoApp.ultimaConsulta;
  
  // Muda de aba
  navegarPara('liquidacao');

  // Preenche processo e partes
  const numInput = document.getElementById('calc-processo-num');
  if (numInput) numInput.value = p.numero || '';

  const partesInput = document.getElementById('calc-partes-nomes');
  if (partesInput) partesInput.value = p.titulo || '';

  // Ajusta indexador de acordo com o tribunal do processo
  const selIndexador = document.getElementById('calc-indexador');
  if (selIndexador && p.tribunal) {
    const trib = p.tribunal.toUpperCase();
    if (trib.includes('TJSP')) selIndexador.value = 'TJSP';
    else if (trib.includes('TJMG')) selIndexador.value = 'TJMG';
    else if (trib.includes('TJRJ')) selIndexador.value = 'TJRJ';
    else if (trib.includes('TRF') || trib.includes('JF')) selIndexador.value = 'JF_GERAL';
    else if (trib.includes('TRT') || trib.includes('TST')) selIndexador.value = 'INPC';
  }

  alert(`Dados do processo ${p.numero} importados com sucesso para a Liquidação Judicial!`);
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
  const btn = document.querySelector('#view-liquidacao .btn-primary');
  const btnOriginalHTML = btn ? btn.innerHTML : '⚡ Processar Liquidação Judicial';
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner"></span> Processando...';
  }

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
  } finally {
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = btnOriginalHTML;
    }
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
  const btn = document.querySelector('#view-revisional .btn-primary');
  const btnOriginalHTML = btn ? btn.innerHTML : '⚡ Calcular Revisão Contratual';
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner"></span> Calculando...';
  }

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
  } finally {
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = btnOriginalHTML;
    }
  }
}

// 7. Prazos Forenses
async function executarCalculoPrazo() {
  const btn = document.querySelector('#view-prazos .btn-primary');
  const btnOriginalHTML = btn ? btn.innerHTML : '⚡ Calcular Vencimento Fatal';
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner"></span> Calculando...';
  }

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
  } finally {
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = btnOriginalHTML;
    }
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
  const btn = document.querySelector('#view-indices .btn-primary') || (typeof event !== 'undefined' && event?.target);
  const btnOriginalHTML = btn ? btn.innerHTML : '🔄 Sincronizar Agora com Bacen/IBGE';
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner"></span> Sincronizando...';
  }

  try {
    const res = await fetch('/api/indices/sincronizar', { method: 'POST' });
    const data = await res.json();
    alert(data.mensagem || "Sincronização com Bacen e IBGE finalizada!");
    carregarIndices();
  } catch (err) {
    alert("Falha na sincronização: " + err.message);
  } finally {
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = btnOriginalHTML;
    }
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
