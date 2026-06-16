class Message {
  final String role; // "user" or "assistant"
  final String content;

  const Message({required this.role, required this.content});

  bool get isUser => role == 'user';

  Map<String, String> toJson() => {'role': role, 'content': content};
}
