import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:speech_to_text/speech_to_text.dart';
import 'package:flutter_tts/flutter_tts.dart';
import '../models/message.dart';
import '../services/chat_service.dart';

class ChatScreen extends StatefulWidget {
  final ChatService service;
  final VoidCallback onToggleTheme;
  final bool isDark;

  const ChatScreen({
    super.key,
    required this.service,
    required this.onToggleTheme,
    required this.isDark,
  });

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final _controller = TextEditingController();
  final _scroll = ScrollController();
  final _messages = <Message>[];
  bool _loading = false;
  bool _cancelar = false; // para el botón de parar

  // Voz
  final _speech = SpeechToText();
  final _tts = FlutterTts();
  bool _speechDisponible = false;
  bool _grabando = false;
  bool _leerEnVoz = false;

  // Personalidad
  String _modo = 'amigable';

  static const _clave = 'mensajes-guardados';
  static const _claveModo = 'modo-asistente';

  static const _modos = {
    'amigable': '😊 Amigable',
    'chapin': '🇬🇹 Chapín',
    'tutor': '🎓 Tutor de tareas',
    'profesor': '👨‍🏫 Profesor',
    'programador': '💻 Programador',
    'gracioso': '😄 Gracioso',
  };

  // [etiqueta, texto que se envía]
  static const _sugerencias = [
    ['📝 Dame un ejemplo', 'Dame un ejemplo'],
    ['🔁 Más simple', 'Explícamelo más simple'],
    ['➡️ ¿Qué más?', '¿Qué más me puedes decir sobre esto?'],
  ];

  static const _preguntas = [
    'Escríbeme un ejemplo de código en Go 💻',
    'Dame una idea para un proyecto',
    'Cuéntame un chiste de programadores 😄',
    '¿Cómo empiezo a aprender a programar?',
  ];

  @override
  void initState() {
    super.initState();
    _cargar();
    _prepararVoz();
  }

  Future<void> _prepararVoz() async {
    final ok = await _speech.initialize(
      onStatus: (s) {
        if ((s == 'done' || s == 'notListening') && mounted) {
          setState(() => _grabando = false);
        }
      },
      onError: (e) {
        if (mounted) setState(() => _grabando = false);
      },
    );
    await _tts.setLanguage('es-ES');
    if (mounted) setState(() => _speechDisponible = ok);
  }

  void _alternarMicro() async {
    if (!_speechDisponible) return;
    if (_grabando) {
      await _speech.stop();
      if (mounted) setState(() => _grabando = false);
      return;
    }
    setState(() => _grabando = true);
    await _speech.listen(
      listenOptions: SpeechListenOptions(localeId: 'es_ES'),
      onResult: (r) {
        setState(() => _controller.text = r.recognizedWords);
      },
    );
  }

  Future<void> _hablar(String texto) async {
    if (!_leerEnVoz) return;
    final limpio = texto
        .replaceAll(RegExp(r'```[\s\S]*?```'), ' código ')
        .replaceAll(RegExp(r'[*`#]'), '');
    await _tts.stop();
    await _tts.speak(limpio);
  }

  @override
  void dispose() {
    _controller.dispose();
    _scroll.dispose();
    _tts.stop();
    super.dispose();
  }

  Future<void> _cargar() async {
    final prefs = await SharedPreferences.getInstance();
    final modo = prefs.getString(_claveModo);
    final guardado = prefs.getString(_clave);
    if (mounted) {
      setState(() {
        if (modo != null && _modos.containsKey(modo)) _modo = modo;
        if (guardado != null) {
          try {
            final lista = (jsonDecode(guardado) as List)
                .map((e) => Message.fromJson(e as Map<String, dynamic>))
                .toList();
            _messages
              ..clear()
              ..addAll(lista);
          } catch (_) {}
        }
      });
    }
  }

  Future<void> _guardar() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(
        _clave, jsonEncode(_messages.map((m) => m.toJson()).toList()));
  }

  Future<void> _guardarModo() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_claveModo, _modo);
  }

  void _bajarDelTodo() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scroll.hasClients) {
        _scroll.animateTo(
          _scroll.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  // ---------- Enviar (streaming con personalidad y botón de parar) ----------
  Future<void> _send(String text) async {
    text = text.trim();
    if (text.isEmpty || _loading) return;

    setState(() {
      _messages.add(Message(role: 'user', content: text));
      _loading = true;
      _cancelar = false;
      _controller.clear();
    });
    _bajarDelTodo();

    final historia = List<Message>.from(_messages);
    final idx = _messages.length;
    setState(() => _messages.add(const Message(role: 'assistant', content: '…')));

    try {
      final buffer = StringBuffer();
      await for (final chunk in widget.service.streamMessages(historia, modo: _modo)) {
        if (_cancelar) break; // el usuario pulsó parar
        buffer.write(chunk);
        if (!mounted) return;
        setState(() => _messages[idx] =
            Message(role: 'assistant', content: buffer.toString()));
        _bajarDelTodo();
      }
      if (buffer.isEmpty && _cancelar) {
        setState(() => _messages.removeAt(idx)); // no llegó nada
      } else {
        _hablar(_messages[idx].content);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _messages[idx] = const Message(
            role: 'assistant',
            content: '⚠️ No me pude conectar, inténtalo de nuevo'));
      }
    } finally {
      if (mounted) setState(() => _loading = false);
      _guardar();
      _bajarDelTodo();
    }
  }

  void _parar() {
    setState(() => _cancelar = true);
  }

  // 🔄 Regenerar la última respuesta.
  void _regenerar() {
    if (_loading || _messages.length < 2) return;
    if (_messages.last.role == 'assistant') _messages.removeLast();
    if (_messages.isEmpty || _messages.last.role != 'user') return;
    final ultimoUser = _messages.removeLast().content;
    setState(() {});
    _send(ultimoUser);
  }

  void _nuevaConversacion() async {
    setState(() {
      _messages.clear();
      _loading = false;
    });
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_clave);
  }

  // 📤 Compartir: copia toda la conversación al portapapeles.
  void _compartir() {
    if (_messages.isEmpty) return;
    final texto = _messages
        .map((m) => (m.isUser ? '🧑 Yo: ' : '🤖 Asistente: ') + m.content)
        .join('\n\n');
    Clipboard.setData(ClipboardData(text: '💬 Conversación con mi Asistente IA\n\n$texto'));
    ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
        content: Text('¡Conversación copiada! Pégala donde quieras 📋'),
        duration: Duration(seconds: 2)));
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final puedeSugerir = !_loading &&
        _messages.isNotEmpty &&
        _messages.last.role == 'assistant' &&
        _messages.last.content != '…' &&
        !_messages.last.content.startsWith('⚠️');

    return Scaffold(
      appBar: AppBar(
        backgroundColor: colors.primary,
        foregroundColor: Colors.white,
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: const [
            Text('Asistente IA',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
            Text('Creado por Samuel',
                style: TextStyle(fontSize: 12, color: Colors.white70)),
          ],
        ),
        actions: [
          // 🎭 Personalidad
          PopupMenuButton<String>(
            tooltip: 'Personalidad',
            icon: const Icon(Icons.theater_comedy),
            onSelected: (m) {
              setState(() => _modo = m);
              _guardarModo();
              ScaffoldMessenger.of(context).showSnackBar(SnackBar(
                  content: Text('Personalidad: ${_modos[m]}'),
                  duration: const Duration(seconds: 1)));
            },
            itemBuilder: (_) => _modos.entries
                .map((e) => PopupMenuItem(
                      value: e.key,
                      child: Text(e.value + (e.key == _modo ? '  ✓' : '')),
                    ))
                .toList(),
          ),
          IconButton(
            tooltip: 'Leer en voz alta',
            icon: Icon(_leerEnVoz ? Icons.volume_up : Icons.volume_off),
            onPressed: () {
              setState(() => _leerEnVoz = !_leerEnVoz);
              if (!_leerEnVoz) _tts.stop();
            },
          ),
          IconButton(
            tooltip: 'Modo claro/oscuro',
            icon: Icon(widget.isDark ? Icons.light_mode : Icons.dark_mode),
            onPressed: widget.onToggleTheme,
          ),
          PopupMenuButton<String>(
            onSelected: (v) {
              if (v == 'nueva') _nuevaConversacion();
              if (v == 'compartir') _compartir();
            },
            itemBuilder: (_) => const [
              PopupMenuItem(value: 'nueva', child: Text('🔄 Nueva conversación')),
              PopupMenuItem(value: 'compartir', child: Text('📤 Compartir')),
            ],
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child:
                _messages.isEmpty ? _bienvenida(colors) : _listaMensajes(colors),
          ),
          if (_loading)
            const Padding(
              padding: EdgeInsets.all(10),
              child: Row(
                children: [
                  SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2)),
                  SizedBox(width: 10),
                  Text('escribiendo…'),
                ],
              ),
            ),
          if (puedeSugerir) _barraSugerencias(colors),
          _barraEscribir(colors),
        ],
      ),
    );
  }

  // 💡 Preguntas sugeridas + 🔄 regenerar.
  Widget _barraSugerencias(ColorScheme colors) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      child: Wrap(
        spacing: 6,
        runSpacing: 4,
        children: [
          ..._sugerencias.map((s) => ActionChip(
                label: Text(s[0], style: const TextStyle(fontSize: 12.5)),
                onPressed: () => _send(s[1]),
              )),
          ActionChip(
            label: const Text('🔄 Otra respuesta', style: TextStyle(fontSize: 12.5)),
            onPressed: _regenerar,
          ),
        ],
      ),
    );
  }

  Widget _bienvenida(ColorScheme colors) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          const SizedBox(height: 20),
          Icon(Icons.smart_toy_outlined, size: 64, color: colors.primary),
          const SizedBox(height: 12),
          const Text('¡Hola! Soy tu asistente.',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600)),
          const SizedBox(height: 4),
          const Text('Prueba con una de estas:',
              style: TextStyle(color: Colors.grey)),
          const SizedBox(height: 16),
          Wrap(
            alignment: WrapAlignment.center,
            spacing: 8,
            runSpacing: 8,
            children: _preguntas
                .map((p) => ActionChip(label: Text(p), onPressed: () => _send(p)))
                .toList(),
          ),
        ],
      ),
    );
  }

  Widget _listaMensajes(ColorScheme colors) {
    return ListView.builder(
      controller: _scroll,
      padding: const EdgeInsets.all(14),
      itemCount: _messages.length,
      itemBuilder: (context, i) {
        final m = _messages[i];
        return Align(
          alignment: m.isUser ? Alignment.centerRight : Alignment.centerLeft,
          child: Container(
            margin: const EdgeInsets.symmetric(vertical: 5),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 11),
            constraints: BoxConstraints(
                maxWidth: MediaQuery.of(context).size.width * 0.8),
            decoration: BoxDecoration(
              color: m.isUser ? colors.primary : colors.surfaceContainerHighest,
              borderRadius: BorderRadius.only(
                topLeft: const Radius.circular(18),
                topRight: const Radius.circular(18),
                bottomLeft: Radius.circular(m.isUser ? 18 : 5),
                bottomRight: Radius.circular(m.isUser ? 5 : 18),
              ),
              boxShadow: [
                BoxShadow(
                    color: Colors.black.withValues(alpha: 0.06),
                    blurRadius: 6,
                    offset: const Offset(0, 2)),
              ],
            ),
            child: m.isUser
                ? Text(m.content,
                    style: const TextStyle(color: Colors.white, height: 1.4))
                : _mensajeAsistente(m.content, colors),
          ),
        );
      },
    );
  }

  // Respuesta del asistente: contenido formateado + botón de copiar (#B).
  Widget _mensajeAsistente(String texto, ColorScheme colors) {
    final esReal = texto != '…' && !texto.startsWith('⚠️');
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        _contenidoAsistente(texto, colors),
        if (esReal)
          Padding(
            padding: const EdgeInsets.only(top: 4),
            child: InkWell(
              onTap: () {
                Clipboard.setData(ClipboardData(text: texto));
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
                    content: Text('¡Respuesta copiada! 📋'),
                    duration: Duration(seconds: 1)));
              },
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.copy, size: 13, color: colors.primary),
                  const SizedBox(width: 4),
                  Text('Copiar',
                      style: TextStyle(fontSize: 12, color: colors.primary)),
                ],
              ),
            ),
          ),
      ],
    );
  }

  Widget _contenidoAsistente(String texto, ColorScheme colors) {
    final partes = texto.split('```');
    final hijos = <Widget>[];
    for (var i = 0; i < partes.length; i++) {
      final parte = partes[i];
      if (i.isOdd) {
        var codigo = parte;
        final salto = codigo.indexOf('\n');
        if (salto != -1) {
          final primera = codigo.substring(0, salto).trim();
          if (!primera.contains(' ') && primera.length < 15) {
            codigo = codigo.substring(salto + 1);
          }
        }
        hijos.add(_bloqueCodigo(codigo.trimRight()));
      } else if (parte.trim().isNotEmpty) {
        hijos.add(Text.rich(
          TextSpan(children: _spans(parte.trim(), colors)),
          style: TextStyle(color: colors.onSurface, height: 1.4),
        ));
      }
    }
    if (hijos.isEmpty) {
      return Text(texto, style: TextStyle(color: colors.onSurface, height: 1.4));
    }
    return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: hijos);
  }

  List<InlineSpan> _spans(String text, ColorScheme colors) {
    final spans = <InlineSpan>[];
    final regex = RegExp(r'\*\*(.+?)\*\*|`(.+?)`');
    int last = 0;
    for (final m in regex.allMatches(text)) {
      if (m.start > last) spans.add(TextSpan(text: text.substring(last, m.start)));
      if (m.group(1) != null) {
        spans.add(TextSpan(
            text: m.group(1),
            style: const TextStyle(fontWeight: FontWeight.bold)));
      } else if (m.group(2) != null) {
        spans.add(TextSpan(
            text: m.group(2),
            style: TextStyle(
                fontFamily: 'monospace',
                backgroundColor: colors.surfaceContainerHighest)));
      }
      last = m.end;
    }
    if (last < text.length) spans.add(TextSpan(text: text.substring(last)));
    return spans;
  }

  Widget _bloqueCodigo(String codigo) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.symmetric(vertical: 6),
      decoration: BoxDecoration(
        color: const Color(0xFF1E1E2E),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Stack(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(14, 34, 14, 14),
            child: SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: SelectableText(
                codigo,
                style: const TextStyle(
                    fontFamily: 'monospace',
                    color: Color(0xFFE7E7F0),
                    fontSize: 13.5),
              ),
            ),
          ),
          Positioned(
            top: 2,
            right: 2,
            child: TextButton.icon(
              onPressed: () {
                Clipboard.setData(ClipboardData(text: codigo));
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
                    content: Text('¡Código copiado!'),
                    duration: Duration(seconds: 1)));
              },
              icon: const Icon(Icons.copy, size: 14, color: Colors.white70),
              label: const Text('Copiar',
                  style: TextStyle(color: Colors.white70, fontSize: 12)),
            ),
          ),
        ],
      ),
    );
  }

  Widget _barraEscribir(ColorScheme colors) {
    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(10, 6, 10, 10),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: _controller,
                textInputAction: TextInputAction.send,
                decoration: InputDecoration(
                  hintText: 'Escribe un mensaje…',
                  filled: true,
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(24),
                    borderSide: BorderSide.none,
                  ),
                ),
                onSubmitted: _send,
              ),
            ),
            if (_speechDisponible) ...[
              const SizedBox(width: 4),
              IconButton(
                tooltip: 'Hablar por voz',
                icon: Icon(_grabando ? Icons.mic : Icons.mic_none,
                    color: _grabando ? Colors.red : colors.primary),
                onPressed: _alternarMicro,
              ),
            ],
            const SizedBox(width: 4),
            // ⏹️ Mientras carga = botón de parar; si no = enviar.
            CircleAvatar(
              backgroundColor: _loading ? Colors.red : colors.primary,
              child: IconButton(
                icon: Icon(_loading ? Icons.stop : Icons.send, color: Colors.white),
                onPressed: _loading ? _parar : () => _send(_controller.text),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
